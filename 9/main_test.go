package main

import (
	"errors"
	"testing"
)

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{
			name: "example",
			in:   "a4bc2d5e",
			want: "aaaabccddddde",
		},
		{
			name: "no digits",
			in:   "abcd",
			want: "abcd",
		},
		{
			name:    "only digits",
			in:      "45",
			want:    "",
			wantErr: ErrInvalidString,
		},
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "multi digit",
			in:   "x12y3",
			want: "xxxxxxxxxxxxyyy",
		},
		{
			name: "unicode runes",
			in:   "ф2я",
			want: "ффя",
		},
		{
			name: "zero",
			in:   "z0a1",
			want: "a",
		},
		{
			name:    "starts with digit",
			in:      "3a",
			want:    "",
			wantErr: ErrInvalidString,
		},
		{
			name: "mixed punctuation",
			in:   "-3a2",
			want: "---aa",
		},
		{
			name: "consecutive letters then number applies to last",
			in:   "ab12",
			want: "abbbbbbbbbbbb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeString(tt.in)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil (out=%q)", got)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("decodeString(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
