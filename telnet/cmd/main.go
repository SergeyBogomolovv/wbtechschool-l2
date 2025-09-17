package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	host, port, timeout := parseFlags(os.Args[1:])

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Fprintln(os.Stderr, "connect timeout exceeded")
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	socketReader := bufio.NewReader(conn)
	socketWriter := bufio.NewWriter(conn)

	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				if _, werr := socketWriter.Write(line); werr != nil {
					fmt.Fprintln(os.Stderr, werr)
					return
				}
				if ferr := socketWriter.Flush(); ferr != nil {
					fmt.Fprintln(os.Stderr, ferr)
					return
				}
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					if tcp, ok := conn.(*net.TCPConn); ok {
						tcp.CloseWrite()
					} else {
						conn.Close()
					}
					return
				}
				fmt.Fprintln(os.Stderr, err)
				return
			}
		}
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		writer := bufio.NewWriter(os.Stdout)
		for {
			data, err := socketReader.ReadBytes('\n')
			if len(data) > 0 {
				if _, werr := writer.Write(data); werr != nil {
					fmt.Fprintln(os.Stderr, werr)
					return
				}
				if ferr := writer.Flush(); ferr != nil {
					fmt.Fprintln(os.Stderr, ferr)
					return
				}
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				fmt.Fprintln(os.Stderr, err)
				return
			}
		}
	}()

	select {
	case <-done:
	case <-sigCh:
	}
}

func parseFlags(args []string) (string, int, time.Duration) {
	fs := flag.NewFlagSet("telnet", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 10*time.Second, "connection timeout")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if fs.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: telnet [--timeout duration] host port")
		os.Exit(2)
	}

	host := fs.Arg(0)
	portStr := fs.Arg(1)
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		fmt.Fprintln(os.Stderr, "invalid port")
		os.Exit(2)
	}

	return host, port, *timeout
}
