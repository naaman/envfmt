package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	InvalidEnvironmentVariableName = errors.New("Invalid characters for environment variable name.")
	fEnvFile                       = flag.String("file", ".env", "/path/to/environment/file")
	fFilter                        = flag.String("filter", "PATH|GIT_DIR|CPATH|CPPATH|LD_PRELOAD|LIBRARY_PATH|PS1", "KEYS|TO|FILTER")
	fInvert                        = flag.Bool("invert", true, "Match the opposite of the filter.")
	filter                         [][]byte
)

func init() {
	flag.Parse()
	for _, f := range []string{*fEnvFile, *fFilter, strconv.FormatBool(*fInvert)} {
		if f == "" {
			flag.Usage()
			os.Exit(1)
		}
	}

	for _, s := range strings.Split(*fFilter, "|") {
		filter = append(filter, []byte(s))
	}
}

func parse(envFile io.Reader, out io.Writer) error {
	r := bufio.NewReader(envFile)

	var ƒ parseStateFunc
	ƒ = new(startLine)
	var k []byte
	var v []byte

	for {
		c, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		ƒ = ƒ.parse(c, &k, &v, out)
	}
}

type parseStateFunc interface {
	parse(c byte, k *[]byte, v *[]byte, w io.Writer) parseStateFunc
}

type startLine struct{}
type readType struct{}
type endType struct{}
type readProc struct{}

func (s *startLine) parse(c byte, k *[]byte, v *[]byte, w io.Writer) parseStateFunc {
	switch {
	case isWhitespace(c):
		return new(startLine)
	case isValid(c):
		*k = append(*k, c)
		return new(readType)
	default:
		panic(InvalidEnvironmentVariableName)
	}
}

func (s *readType) parse(c byte, k *[]byte, v *[]byte, w io.Writer) parseStateFunc {
	switch {
	case isValid(c):
		*k = append(*k, c)
		return new(readType)
	case c == '=':
		return new(readProc)
	case isIgnored(c):
		return new(endType)
	default:
		w.Write([]byte{c})
		panic(InvalidEnvironmentVariableName)
	}
}

func (s *endType) parse(c byte, k *[]byte, v *[]byte, w io.Writer) parseStateFunc {
	switch {
	case isIgnored(c):
		return new(endType)
	case c == '=':
		return new(readProc)
	default:
		panic(InvalidEnvironmentVariableName)
	}
}

func (s *readProc) parse(c byte, k *[]byte, v *[]byte, w io.Writer) parseStateFunc {
	switch c {
	case 0:
		if !isFiltered(k) {
			w.Write(*k)
			w.Write([]byte{'='})
			w.Write(*v)
			w.Write([]byte{'\n'})
		}
		*k = nil
		*v = nil
		return new(startLine)
	case '\n':
		*v = append(*v, '\\', 'n')
		return new(readProc)
	default:
		*v = append(*v, c)
		return new(readProc)
	}
}

func isFiltered(k *[]byte) bool {
	for _, kFiltered := range filter {
		if bytes.Equal(kFiltered, *k) {
			return true
		}
	}
	return false
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

func isIgnored(c byte) bool {
	return c == ' ' || c == '\t'
}

// http://pubs.opengroup.org/onlinepubs/000095399/basedefs/xbd_chap08.html
func isValid(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c == '_')
}

func main() {
	envFile, _ := os.Open(*fEnvFile)
	parse(envFile, os.Stdout)
}
