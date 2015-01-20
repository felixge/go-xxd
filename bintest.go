package main

import "fmt"

func main() {
	a := "01101000" // 01100101 01101100 01101100 01101111 00101100 00100000 01110111 01101111 01110010 01101100 01100100 00100001"

	_, foo := decode([]byte(a))

	fmt.Println(string(foo))
}

func decode(src []byte) (int, byte) {
	var (
		d byte
		b byte
		k int
	)

	for i, v := range src {
		d *= 2
		b <<= 1
		fmt.Println(d, b)
		if v == 32 {
			k = i
		}
		if v == '1' {
			d += 1
			b ^= 1
		}
		fmt.Println(b, d)
	}
	return k, d
}
