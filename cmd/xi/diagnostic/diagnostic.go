package diagnostic

import (
	"bufio"
	"errors"
	"fmt"
	goAst "go/ast"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/manapointer/xi/pkg/parser"
	"github.com/manapointer/xi/pkg/scanner"
	"github.com/manapointer/xi/pkg/token"
	"github.com/spf13/cobra"
)

type diagnosticOptions struct {
	lex   bool
	parse bool
	trace bool
}

func NewDiagnosticCmd() *cobra.Command {
	opts := &diagnosticOptions{}

	cmd := &cobra.Command{
		Use:   "diagnostic [diagnostic flags] [files]",
		Short: "Diagnostic outputs diagnostic information for various stages of the Xi compiler.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run(args)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.lex, "lex", false, "Output lexing information")
	flags.BoolVar(&opts.parse, "parse", false, "Output parsing information")
	flags.BoolVar(&opts.trace, "trace", false, "Trace parsing")

	return cmd
}

func (opts *diagnosticOptions) run(files []string) error {
	switch {
	case opts.lex:
		return opts.runLex(files)
	case opts.parse:
		return opts.runParse(files)
	}

	return nil
}

func (opts *diagnosticOptions) runLex(files []string) error {
	for _, file := range files {
		f, err := openDiagnosticFile(file, ".lexed")
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		defer w.Flush()

		src, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		s := scanner.NewScanner(src, nil)
		for tok := s.Scan(); tok.Typ != token.Eof; tok = s.Scan() {
			fprintTokenDiagnostic(w, tok)
			if tok.Typ == token.Error {
				return errors.New(tok.Lit)
			}
		}

		w.Flush()
	}

	return nil
}

func (opts *diagnosticOptions) mode() (mode parser.Mode) {
	if opts.trace {
		mode |= parser.Trace
	}

	return
}

func (opts *diagnosticOptions) runParse(files []string) error {
	for _, file := range files {
		f, err := openDiagnosticFile(file, ".parsed")
		if err != nil {
			return err
		}
		defer f.Close()

		astf, err := parser.ParseFile(file, nil, opts.mode())
		if err != nil {
			fmt.Fprint(f, err)
			return err
		}

		err = goAst.Fprint(f, nil, astf, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func openDiagnosticFile(filename, suffix string) (*os.File, error) {
	dir := path.Dir(filename)
	base := path.Base(filename)
	trimmed := strings.TrimSuffix(base, path.Ext(base))

	f, err := os.OpenFile(path.Join(dir, trimmed+suffix), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func fprintTokenDiagnostic(w io.Writer, tok token.Token) {
	repr := tok.Lit

	switch tok.Typ {
	case token.Integer:
		repr = "integer " + tok.Lit
	case token.String:
		repr = "string " + tok.Lit[1:len(tok.Lit)-1]
	case token.Char:
		repr = "character " + tok.Lit[1:len(tok.Lit)-1]
	case token.Ident:
		repr = "id " + tok.Lit
	}

	fmt.Fprintf(w, "%d:%d %s\n", tok.Pos.Line, tok.Pos.Column, repr)
}
