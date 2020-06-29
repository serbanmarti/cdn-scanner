package cli

import (
	"github.com/spf13/cobra"
)

// ServiceCmd interface for all command that will be written
type ServiceCmd interface {
	GetCmd() *cobra.Command
	Setup() error
	Execute() error
}

// Service structure for all command that will be written
type Service struct {
	cmd *cobra.Command
}

// NewCmd returns a new cobra sub-command for the root and runs it when called
func NewCmd(serviceCmd ServiceCmd) *cobra.Command {
	cmd := serviceCmd.GetCmd()

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := serviceCmd.Setup(); err != nil {
			return err
		}
		return serviceCmd.Execute()
	}

	return cmd
}
