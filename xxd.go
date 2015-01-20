package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	flag "github.com/ogier/pflag"
)

// usage and version
const (
	Help = `Usage:
       xxd [options] [infile [outfile]]
    or
       xxd -r [-s offset] [-c cols] [--ps] [infile [outfile]]
Options:
    -a, --autoskip     toggle autoskip: A single '*' replaces nul-lines. Default off.
    -b, --binary       binary digit dump (incompatible with -ps,-i,-r). Default hex.
    -c, --cols         format <cols> octets per line. Default 16 (-i 12, --ps 30).
    -E, --ebcdic       show characters in EBCDIC. Default ASCII.
    -g, --groups       number of octets per group in normal output. Default 2.
    -h, --help         print this summary.
    -i, --include      output in C include file style.
    -l, --length       stop after <len> octets.
        --ps           output in postscript plain hexdump style.
    -r, --reverse      reverse operation: convert (or patch) hexdump into ASCII output.
    -s, --seek         start at <seek> bytes in file.
    -u, --uppercase    use upper case hex letters.
    -v, --version      show version.`
	Version = `xxd v2.0 2014-17-01 by Felix Geisendörfer and Eric Lagergren`
)

// cli flags
var (
	autoskip   = flag.BoolP("autoskip", "a", false, "toggle autoskip (* replaces nul lines")
	binary     = flag.BoolP("binary", "b", false, "binary dump, incompatible with -ps, -i, -r")
	columns    = flag.IntP("cols", "c", -1, "format <cols> octets per line")
	ebcdic     = flag.BoolP("ebcdic", "E", false, "use EBCDIC instead of ASCII")
	group      = flag.IntP("group", "g", -1, "num of octets per group")
	cfmt       = flag.BoolP("include", "i", false, "output in C include format")
	length     = flag.Int64P("len", "l", -1, "stop after len octets")
	postscript = flag.Bool("ps", false, "output in postscript plain hd style")
	reverse    = flag.BoolP("reverse", "r", false, "convert hex to binary")
	offset     = flag.Int("off", 0, "revert with offset")
	seek       = flag.Int64P("seek", "s", 0, "start at seek bytes abs")
	upper      = flag.BoolP("uppercase", "u", false, "use uppercase hex letters")
	version    = flag.BoolP("version", "v", false, "print version")
)

// constants used in xxd()
const (
	ebcdicOffset = 0x40
)

// dumpType enum
const (
	dumpHex = iota
	dumpBinary
	dumpCformat
	dumpPostscript
)

// variables used in xxd*()
var (
	dumpType int

	space        = []byte(" ")
	doubleSpace  = []byte("  ")
	dot          = []byte(".")
	newLine      = []byte("\n")
	zeroHeader   = []byte("0000000: ")
	unsignedChar = []byte("unsigned char ")
	unsignedInt  = []byte("};\nunsigned int ")
	lenEquals    = []byte("_len = ")
	brackets     = []byte("[] = {")
	asterisk     = []byte("*")
	hexPrefix    = []byte("0x")
	commaSpace   = []byte(", ")
	comma        = []byte(",")
	semiColonNl  = []byte(";\n")
)

// ascii -> ebcdic lookup table
var ebcdicTable = []byte{
	0040, 0240, 0241, 0242, 0243, 0244, 0245, 0246,
	0247, 0250, 0325, 0056, 0074, 0050, 0053, 0174,
	0046, 0251, 0252, 0253, 0254, 0255, 0256, 0257,
	0260, 0261, 0041, 0044, 0052, 0051, 0073, 0176,
	0055, 0057, 0262, 0263, 0264, 0265, 0266, 0267,
	0270, 0271, 0313, 0054, 0045, 0137, 0076, 0077,
	0272, 0273, 0274, 0275, 0276, 0277, 0300, 0301,
	0302, 0140, 0072, 0043, 0100, 0047, 0075, 0042,
	0303, 0141, 0142, 0143, 0144, 0145, 0146, 0147,
	0150, 0151, 0304, 0305, 0306, 0307, 0310, 0311,
	0312, 0152, 0153, 0154, 0155, 0156, 0157, 0160,
	0161, 0162, 0136, 0314, 0315, 0316, 0317, 0320,
	0321, 0345, 0163, 0164, 0165, 0166, 0167, 0170,
	0171, 0172, 0322, 0323, 0324, 0133, 0326, 0327,
	0330, 0331, 0332, 0333, 0334, 0335, 0336, 0337,
	0340, 0341, 0342, 0343, 0344, 0135, 0346, 0347,
	0173, 0101, 0102, 0103, 0104, 0105, 0106, 0107,
	0110, 0111, 0350, 0351, 0352, 0353, 0354, 0355,
	0175, 0112, 0113, 0114, 0115, 0116, 0117, 0120,
	0121, 0122, 0356, 0357, 0360, 0361, 0362, 0363,
	0134, 0237, 0123, 0124, 0125, 0126, 0127, 0130,
	0131, 0132, 0364, 0365, 0366, 0367, 0370, 0371,
	0060, 0061, 0062, 0063, 0064, 0065, 0066, 0067,
	0070, 0071, 0372, 0373, 0374, 0375, 0376, 0377,
}

func cfmtEncode(dst, src []byte, hextable string) {
	dst[0] = '0'
	dst[1] = 'x'
	for i, v := range src {
		dst[i+1*2] = hextable[v>>4]
		dst[i+1*2+1] = hextable[v&0x0f]
	}
}

// convert a byte into its binary representation
func binaryEncode(dst, src []byte) {
	d := uint(0)
	for i := 7; i >= 0; i-- {
		if src[0]&(1<<d) == 0 {
			dst[i] = '0'
		} else {
			dst[i] = '1'
		}
		d++
	}
}

// returns -1 on success
// returns k > -1 if space found where k is index of space byte
func binaryDecode(dst, src []byte) int {
	var d byte

	for i, v := range src {
		d <<= 1
		if isSpace(&v) { // found a space, so between groups
			if i == 0 {
				return 1
			}
			return i
		}
		if v == '1' {
			d ^= 1
		}
	}
	if d < 32 || d > 127 {
		return 1
	}

	dst[0] = d
	return -1
}

// hex lookup table for hexEncode()
const (
	ldigits = "0123456789abcdef"
	udigits = "0123456789ABCDEF"
)

// copied from encoding/hex package in order to add support for uppercase hex
func hexEncode(dst, src []byte, hextable string) {
	for i, v := range src {
		dst[i*2] = hextable[v>>4]
		dst[i*2+1] = hextable[v&0x0f]
	}
}

// copied from encoding/hex package
// returns -1 on bad byte
// returns -2 on space (\n, \t, \s)
// returns -3 on two consecutive spaces
// returns 0 on success
func hexDecode(dst, src []byte) int {
	if isSpace(&src[0]) {
		if isSpace(&src[1]) {
			return -3
		}
		return -2
	}

	for i := 0; i < len(src)/2; i++ {
		a, ok := fromHexChar(src[i*2])
		if !ok {
			return -1
		}
		b, ok := fromHexChar(src[i*2+1])
		if !ok {
			return -1
		}

		// check bounds
		r := (a << 4) | b
		if r < 32 || r > 127 {
			return -3
		} else {
			dst[0] = r
		}
	}
	return 0
}

// copied from encoding/hex package
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}

// check if entire line is full of empty []byte{0} bytes (nul in C)
func empty(b *[]byte) bool {
	for _, v := range *b {
		if v != 0 {
			return false
		}
	}
	return true
}

func isSpace(b *byte) bool {
	if *b == 32 ||
		*b == 9 ||
		*b == 12 {
		return true
	}
	return false
}

func xxdReverse(r io.Reader, w io.Writer) error {
	var (
		cols int
		octs int
		char = make([]byte, 1)
	)

	if *columns != -1 {
		cols = *columns
	}

	switch dumpType {
	case dumpBinary:
		octs = 8
	case dumpPostscript:
		octs = 0
	case dumpCformat:
		octs = 4
	default:
		octs = 2
	}

	if *length != -1 {
		if *length < int64(cols) {
			cols = int(*length)
		}
	}

	if octs < 1 {
		octs = cols
	}

	c := int64(0) // number of characters
	rd := bufio.NewReader(r)
	for {
		line, err := rd.ReadBytes('\n') // read up until a newline
		n := len(line)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}

		if n == 0 {
			return nil
		}

		if dumpType == dumpHex {
			// skip first 8 of line because it's the counter thingy
			// don't go to EOL because COLS bytes of that is the human-
			// readable output
			for len(line) >= 2 {
				if n := hexDecode(char, line[0:2]); n == 0 {
					line = line[2:]
					w.Write(char)
					c++
				} else if n == -1 || n == -2 {
					line = line[1:]
				} else if n == -3 {
					line = line[2:]
				}
			}
		} else if dumpType == dumpBinary {
			for len(line) >= 8 {
				if n := binaryDecode(char, line[0:8]); n != -1 {
					line = line[1:]
					continue
				} else {
					w.Write(char)
					line = line[8:]
					c++
				}
			}
		}

		// For some reason "xxd FILE | xxd -r -c N" truncates the output,
		// so we'll do it as well
		// "xxd FILE | xxd -r -l N" doesn't truncate
		if c == int64(cols) {
			return nil
		}
	}
	return nil
}

func xxd(r io.Reader, w io.Writer, fname string) error {
	var (
		lineOffset int64
		hexOffset  = make([]byte, 6)
		groupSize  int
		cols       int
		octs       int
		caps       = ldigits
		doCHeader  = true
		doCEnd     bool
		// enough room for "unsigned char NAME_FORMAT[] = {"
		varDeclChar = make([]byte, 14+len(fname)+6)
		// enough room for "unsigned int NAME_FORMAT = "
		varDeclInt = make([]byte, 16+len(fname)+7)
		nulLine    int64
		totalOcts  int64
	)

	// Generate the first and last line in the -i output:
	// e.g. unsigned char foo_txt[] = { and unsigned int foo_txt_len =
	if dumpType == dumpCformat {
		// copy over "unnsigned char " and "unsigned int"
		_ = copy(varDeclChar[0:14], unsignedChar[:])
		_ = copy(varDeclInt[0:16], unsignedInt[:])

		for i := 0; i < len(fname); i++ {
			if fname[i] != '.' {
				varDeclChar[14+i] = fname[i]
				varDeclInt[16+i] = fname[i]
			} else {
				varDeclChar[14+i] = '_'
				varDeclInt[16+i] = '_'
			}
		}
		// copy over "[] = {" and "_len = "
		_ = copy(varDeclChar[14+len(fname):], brackets[:])
		_ = copy(varDeclInt[16+len(fname):], lenEquals[:])
	}

	// Switch between upper- and lower-case hex chars
	if *upper {
		caps = udigits
	}

	// xxd -bpi FILE outputs in binary format
	// xxd -b -p -i FILE outputs in C format
	// simply catch the last option since that's what I assume the author
	// wanted...
	if *columns == -1 {
		switch dumpType {
		case dumpPostscript:
			cols = 30
		case dumpCformat:
			cols = 12
		case dumpBinary:
			cols = 6
		default:
			cols = 16
		}
	} else {
		cols = *columns
	}

	// See above comment
	switch dumpType {
	case dumpBinary:
		octs = 8
		groupSize = 1
	case dumpPostscript:
		octs = 0
	case dumpCformat:
		octs = 4
	default:
		octs = 2
		groupSize = 2
	}

	if *group != -1 {
		groupSize = *group
	}

	// If -l is smaller than the number of cols just truncate the cols
	if *length != -1 {
		if *length < int64(cols) {
			cols = int(*length)
		}
	}

	if octs < 1 {
		octs = cols
	}

	// These are bumped down from the beginning of the function in order to
	// allow for their sizes to be allocated based on the user's speficiations
	var (
		line = make([]byte, cols)
		char = make([]byte, octs)
	)

	c := int64(0) // number of characters
	nl := int64(0)
	r = bufio.NewReader(r)
	for {
		n, err := io.ReadFull(r, line)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}

		// Speed it up a bit ;)
		if dumpType == dumpPostscript && n != 0 {
			// Post script values
			// Basically just raw hex output
			for i := 0; i < n; i++ {
				hexEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++
			}
			continue
		}

		if n == 0 {
			if dumpType == dumpPostscript {
				w.Write(newLine)
			}
			return nil // Hidden return!
		} else if n == 0 && dumpType == dumpCformat {
			doCEnd = true
		}

		if *length != -1 {
			if totalOcts == *length {
				break
			}
			totalOcts += *length
		}

		if *autoskip && empty(&line) {
			if nulLine == 1 {
				w.Write(asterisk)
				w.Write(newLine)
			}

			nulLine++

			if nulLine > 1 {
				lineOffset++ // continue to increment our offset
				continue
			}
		}

		if dumpType <= dumpBinary { // either hex or binary
			// Line offset
			hexOffset = strconv.AppendInt(hexOffset[0:0], lineOffset, 16)
			w.Write(zeroHeader[0:(6 - len(hexOffset))])
			w.Write(hexOffset)
			w.Write(zeroHeader[6:])
			lineOffset++
		} else if doCHeader {
			w.Write(varDeclChar)
			w.Write(newLine)
			doCHeader = false
		}

		if dumpType == dumpBinary {
			// Binary values
			for i, k := 0, octs; i < n; i, k = i+1, k+octs {
				binaryEncode(char, line[i:i+1])
				w.Write(char)
				c++

				if k == octs*groupSize {
					k = 0
					w.Write(space)
				}
			}
		} else if dumpType == dumpCformat {
			// C values
			if !doCEnd {
				w.Write(doubleSpace)
			}
			for i := 0; i < n; i++ {
				cfmtEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++

				// don't add spaces to EOL
				if i != n-1 {
					w.Write(commaSpace)
				} else if doCEnd {
					w.Write(comma)
				}
			}
		} else {
			// Hex values -- default xxd FILE output
			for i, k := 0, octs; i < n; i, k = i+1, k+octs {
				hexEncode(char, line[i:i+1], caps)
				w.Write(char)
				c++

				if k == octs*groupSize {
					k = 0 // reset counter
					w.Write(space)
				}
			}
		}

		if doCEnd {
			w.Write(varDeclInt)
			w.Write([]byte(strconv.FormatInt(c, 10)))
			w.Write(semiColonNl)
			return nil
		}

		if n < len(line) && dumpType <= dumpBinary {
			for i := n * octs; i < len(line)*octs; i++ {
				w.Write(space)

				if i%octs == 1 {
					w.Write(space)
				}
			}
		}

		if dumpType != dumpCformat {
			w.Write(space)
		}

		if dumpType <= dumpBinary {
			// Character values
			b := line[:n]
			// EBCDIC
			if *ebcdic {
				for _, c := range b {
					if c >= ebcdicOffset {
						e := ebcdicTable[c-ebcdicOffset : c-ebcdicOffset+1]
						if e[0] > 0x1f && e[0] < 0x7f {
							w.Write(e)
						} else {
							w.Write(dot)
						}
					} else {
						w.Write(dot)
					}
				}
				// ASCII
			} else {
				for i, c := range b {
					if c > 0x1f && c < 0x7f {
						w.Write(line[i : i+1])
					} else {
						w.Write(dot)
					}
				}
			}
		}
		w.Write(newLine)
		nl++
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", Help)
		os.Exit(0)
	}
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stderr, "%s\n", Version)
		os.Exit(0)
	}

	if flag.NArg() > 2 {
		log.Fatalf("too many arguments after %s\n", flag.Args()[1])
	}

	var (
		err  error
		file string
	)

	if flag.NArg() >= 1 {
		file = flag.Args()[0]
	} else {
		file = "-"
	}

	var inFile *os.File
	if file == "-" {
		inFile = os.Stdin
		file = "stdin"
	} else {
		inFile, err = os.Open(file)
		if err != nil {
			log.Fatalln(err)
		}
	}
	defer inFile.Close()

	// Start *seek bytes into file
	if *seek != 0 {
		_, err = inFile.Seek(*seek, os.SEEK_SET)
		if err != nil {
			log.Fatalln(err)
		}
	}

	var outFile *os.File
	if flag.NArg() == 2 {
		outFile, err = os.Open(flag.Args()[1])
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		outFile = os.Stdout
	}
	defer outFile.Close()

	switch true {
	case *binary:
		dumpType = dumpBinary
	case *cfmt:
		dumpType = dumpCformat
	case *postscript:
		dumpType = dumpPostscript
	default:
		dumpType = dumpHex
	}

	out := bufio.NewWriter(outFile)
	defer out.Flush()

	if *reverse {
		if err := xxdReverse(inFile, out); err != nil {
			log.Fatalln(err)
		}
		return
	} else {
		if err := xxd(inFile, out, file); err != nil {
			log.Fatalln(err)
		}
		return
	}
}
