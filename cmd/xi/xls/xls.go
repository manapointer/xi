package xls

import (
	"fmt"

	"github.com/spf13/cobra"
)

type xlsOptions struct{}

func NewXlsCommand() *cobra.Command {
	opts := &xlsOptions{}

	cmd := &cobra.Command{
		Use:   "xls",
		Short: "xls is an implementation of the Language Server Protocol for Xi.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run()
		},
	}

	return cmd
}

func (opts *xlsOptions) run() error {
	fmt.Println("Welcome to xls, the Xi language server!")
	return nil
}
