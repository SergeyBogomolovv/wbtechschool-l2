package sort

import (
	"bytes"
	"fmt"
	"io"
)

type uniqueWriter struct {
	dst  io.Writer
	seen map[string]struct{}
	buf  bytes.Buffer
}

func newUniqueWriter(w io.Writer) *uniqueWriter {
	return &uniqueWriter{dst: w, seen: make(map[string]struct{})}
}

func (u *uniqueWriter) Write(p []byte) (int, error) {
	n := len(p)

	if _, err := u.buf.Write(p); err != nil {
		return 0, err
	}

	for {
		all := u.buf.Bytes()
		i := bytes.IndexByte(all, '\n')
		if i == -1 {
			break
		}

		line := all[:i]
		line = bytes.TrimRight(line, "\r")
		key := string(line)

		u.buf.Next(i + 1)

		if _, seen := u.seen[key]; !seen {
			u.seen[key] = struct{}{}
			if _, err := fmt.Fprintln(u.dst, key); err != nil {
				return 0, err
			}
		}
	}

	return n, nil
}

func (u *uniqueWriter) Flush() error {
	if u.buf.Len() == 0 {
		return nil
	}
	line := bytes.TrimRight(u.buf.Bytes(), "\r")
	key := string(line)
	u.buf.Reset()
	if _, seen := u.seen[key]; !seen {
		u.seen[key] = struct{}{}
		_, err := fmt.Fprintln(u.dst, key)
		return err
	}
	return nil
}
