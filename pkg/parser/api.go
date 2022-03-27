package parser

import (
	"io"
	"io/ioutil"

	"github.com/manapointer/xi/pkg/ast"
)

type Mode int

const (
	Trace Mode = (1 << iota)
)

func readSource(filename string, src interface{}) ([]byte, error) {
	switch t := src.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	case io.Reader:
		return ioutil.ReadAll(t)
	}

	return ioutil.ReadFile(filename)
}

func ParseFile(filename string, src interface{}, mode Mode) (file *ast.File, err error) {
	_, err = readSource(filename, src)
	if err != nil {
		return nil, err
	}

	content, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			switch t := e.(type) {
			case error:
				file = nil
				err = t
			default:
				panic(e)
			}
		}
	}()

	p.init(filename, content, mode)

	f := p.parseFile()
	return f, nil
}
