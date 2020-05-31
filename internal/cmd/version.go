package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/rafaelespinoza/notexfr/internal/version"
)

var _Version = func(cmdName string) directive {
	return &command{
		description: "metadata about the build",
		setup: func(a *arguments) *flag.FlagSet {
			flags := flag.NewFlagSet(cmdName, flag.ExitOnError)
			flags.Usage = func() {
				fmt.Printf(`Usage: %s %s

Description:

	Shows info about the build.
`, _Bin, cmdName)
			}
			return flags
		},
		run: func(ctx context.Context, a *arguments) error {
			version.Show()
			return nil
		},
	}
}("version")
