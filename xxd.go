package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

const (
	byteOffsetInit = 8
	charOffsetInt  = 39
	line_length    = 50
)

func main() {
	line_offset := 0

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [file]\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	buf := make([]byte, 16)
	for {
		n, err := r.Read(buf)
		if n == 0 || err == io.EOF {
			break
		}

		// Line offset
		fmt.Printf("%06x0: ", line_offset)
		line_offset++

		// Hex values
		for i := 0; i < n; i++ {
			fmt.Printf("%02x", buf[i])

			if i%2 == 1 {
				fmt.Print(" ")
			}
		}
		if n < len(buf) {
			for i := n; i < len(buf); i++ {
				fmt.Printf("  ")
				if i%2 == 1 {
					fmt.Print(" ")
				}
			}
		}

		fmt.Printf(" ")

		// Character values
		b := buf[:n]
		for len(b) > 0 {
			r, size := utf8.DecodeRune(b)

			if int(r) > 0x1f && int(r) < 0x7f {
				fmt.Printf("%v", string(r))
			} else {
				fmt.Printf(".")
			}
			b = b[size:]
		}

		fmt.Printf("\n")
	}
}
