package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	if err := XXD(f, out); err != nil {
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
	hexChar := make([]byte, 2)
	for {
		n, err := io.ReadFull(r, buf)
		if n == 0 || err == io.EOF {
			break
		}

		// Line offset
		fmt.Fprintf(w, "%06x0: ", line_offset)
		line_offset++

		// Hex values
		for i := 0; i < n; i++ {
			hex.Encode(hexChar, buf[i:i+1])
			w.Write(hexChar)

			if i%2 == 1 {
				io.WriteString(w, " ")
			}
		}
		if n < len(buf) {
			for i := n; i < len(buf); i++ {
				io.WriteString(w, "  ")
				if i%2 == 1 {
					io.WriteString(w, " ")
				}
			}
		}

		io.WriteString(w, " ")

		// Character values
		b := buf[:n]
		for _, c := range b {
			if c > 0x1f && c < 0x7f {
				io.WriteString(w, string(c))
			} else {
				io.WriteString(w, ".")
			}
		}

		io.WriteString(w, "\n")
	}
	return nil
}
