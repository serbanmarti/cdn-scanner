package cli

import (
	"os"

	"github.com/spf13/cobra"
)

// Execute the command line application
func Execute() {
	root := makeRootCommand()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// makeRootCommand creates the root Cobra command and adds all needed sub-commands
func makeRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "cdn-scanner",
	}

	root.AddCommand(NewCmd(NewScanCmd()))

	return root
}
