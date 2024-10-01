package cmd

import (
	"github.com/spf13/cobra"

	"github.com/rafaelespinoza/notexfr/internal/version"
)

func makeVersion(cmdName string) *cobra.Command {
	return &cobra.Command{
		Use:   cmdName,
		Short: "metadata about the build",
		RunE: func(cmd *cobra.Command, args []string) error {
			version.Show()
			return nil
		},
	}
}
