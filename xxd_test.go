package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
	"testing/quick"
)

var xxdFile = flag.String("xxdFile", "", "File to test against.")

func TestXXD(t *testing.T) {
	if *xxdFile == "" {
		t.Skip("-xxdFile argument not given")
	}
	data, err := ioutil.ReadFile(*xxdFile)
	if err != nil {
		t.Fatal(err)
	}
	test := func(fn func(r io.Reader, w io.Writer) error) func(n uint64) []string {
		return func(n uint64) []string {
			size := n % uint64(len(data))
			fmt.Printf("%d\n", size)
			var out bytes.Buffer
			if err := fn(bytes.NewBuffer(data[0:size]), &out); err != nil {
				return []string{err.Error()}
			}
			return strings.Split(out.String(), "\n")
		}
	}
	if err := quick.CheckEqual(test(XXD), test(xxdNative), nil); err != nil {
		cErr := err.(*quick.CheckEqualError)
		size := cErr.In[0].(uint64) % uint64(len(data))
		for i := range cErr.Out1[0].([]string) {
			got := cErr.Out1[0].([]string)[i]
			want := cErr.Out2[0].([]string)[i]
			if got != want {
				t.Errorf("size: %d\n\ngot : %s\nwant: %s\n", size, got, want)
				break
			}
		}
	}
}

func xxdNative(r io.Reader, w io.Writer) error {
	xxd := exec.Command("xxd", "-")
	xxd.Stdin = r
	xxd.Stdout = w
	xxd.Stderr = w
	return xxd.Run()
}
