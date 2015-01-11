package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
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

var (
	space       = []byte(" ")
	doubleSpace = []byte("  ")
	dot         = []byte(".")
	newline     = []byte("\n")
)

func XXD(r io.Reader, w io.Writer) error {
	var line_offset int64

	r = bufio.NewReader(r)
	buf := make([]byte, 16)
	hexChar := make([]byte, 2)
	zeroHeader := []byte("0000000: ")
	hexOffset := make([]byte, 6)
	for {
		n, err := io.ReadFull(r, buf)
		if n == 0 || err == io.EOF {
			break
		}

		// Line offset
		hexOffset = strconv.AppendInt(hexOffset[0:0], line_offset, 16)
		w.Write(zeroHeader[0:(6 - len(hexOffset))])
		w.Write(hexOffset)
		w.Write(zeroHeader[6:])
		line_offset++

		// Hex values
		for i := 0; i < n; i++ {
			hex.Encode(hexChar, buf[i:i+1])
			w.Write(hexChar)

			if i%2 == 1 {
				w.Write(space)
			}
		}
		if n < len(buf) {
			for i := n; i < len(buf); i++ {
				w.Write(doubleSpace)
				if i%2 == 1 {
					w.Write(space)
				}
			}
		}

		w.Write(space)

		// Character values
		b := buf[:n]
		for i, c := range b {
			if c > 0x1f && c < 0x7f {
				w.Write(buf[i : i+1])
			} else {
				w.Write(dot)
			}
		}

		w.Write(newline)
	}
	return nil
}
