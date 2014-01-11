package main

import (
	"bufio"
	"io"
	"errors"
	"os"
)

var (
	InvalidEnvironmentVariableName = errors.New("Invalid characters for environment variable name.")
)

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
		w.Write([]byte{'\n'})
		w.Write(*k)
		w.Write([]byte{'='})
		w.Write(*v)
		*k = nil
		*v = nil
		return new(startLine)
	default:
		*v = append(*v, c)
		return new(readProc)
	}
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

func isIgnored(c byte) bool {
	return c == ' ' || c == '\t'
}

// http://pubs.opengroup.org/onlinepubs/000095399/basedefs/xbd_chap08.html
func isValid(c byte) bool {
	return (c > '0' && c < '9') ||
		(c > 'A' && c < 'Z') ||
		(c > 'a' && c < 'z') ||
		(c == '_')
}

func main() {
	envFilePath := os.Args[1]
	envFile, _ := os.Open(envFilePath)
	parse(envFile, os.Stdout)
}
