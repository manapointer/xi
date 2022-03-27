package main

import (
	"log"

	"github.com/manapointer/xi/cmd/xi/diagnostic"
	"github.com/spf13/cobra"
)

func main() {
	root := newRootCommand()

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Short:         "Xi does everything related to your Xi source code!",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(
		diagnostic.NewDiagnosticCmd(),
	)

	return cmd
}
