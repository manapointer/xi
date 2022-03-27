package parser

import (
	"io"
	"io/ioutil"

	"github.com/manapointer/xi/pkg/ast"
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

func ParseFile(filename string, src interface{}) (*ast.File, error) {
	_, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	content, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	p.init(filename, content)

	f := p.parseFile()
	return f, nil
}
