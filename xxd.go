package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [file]\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := XXD(f, os.Stdout); err != nil {
		panic(err)
	}
}

const (
	byteOffsetInit = 8
	charOffsetInt  = 39
	line_length    = 50
)

func XXD(r io.Reader, w io.Writer) error {
	line_offset := 0

	r = bufio.NewReader(r)
	buf := make([]byte, 16)
	for {
		n, err := r.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}

		// Line offset
		fmt.Fprintf(w, "%06x0: ", line_offset)
		line_offset++

		// Hex values
		for i := 0; i < n; i++ {
			fmt.Fprintf(w, "%02x", buf[i])

			if i%2 == 1 {
				fmt.Fprint(w, " ")
			}
		}
		if n < len(buf) {
			for i := n; i < len(buf); i++ {
				fmt.Fprintf(w, "  ")
				if i%2 == 1 {
					fmt.Fprint(w, " ")
				}
			}
		}

		fmt.Fprintf(w, " ")

		// Character values
		b := buf[:n]
		for len(b) > 0 {
			r, size := utf8.DecodeRune(b)

			if int(r) > 0x1f && int(r) < 0x7f {
				fmt.Fprintf(w, "%v", string(r))
			} else {
				fmt.Fprintf(w, ".")
			}
			b = b[size:]
		}

		fmt.Fprintf(w, "\n")
	}
	return nil
}
