package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
	"github.com/rafaelespinoza/notexfr/internal/repo/enex"
)

var _Enex = func(cmdName string) *delegator {
	cmd := &delegator{description: "do enex stuff"}
	const helpLink = "https://help.evernote.com/hc/en-us/articles/209005557"

	// Define this field before defining the Usage function so the subcommand
	// descriptions are present in the help message.
	cmd.subs = map[string]directive{
		"to-json": &command{
			description: "convert ENEX file to JSON",
			setup: func(a *arguments) *flag.FlagSet {
				subcmdName := cmdName + " to-json"
				flags := flag.NewFlagSet(subcmdName, flag.ExitOnError)
				var opts interactor.FetchWriteOptions
				flags.StringVar(
					&opts.InputFilename,
					"input",
					"",
					"path to evernote export file",
				)
				flags.StringVar(
					&opts.OutputFilename,
					"output",
					"",
					"path to write data as JSON",
				)
				var timeout time.Duration
				flags.DurationVar(
					&timeout,
					"timeout",
					time.Duration(15)*time.Second,
					"how long to wait before timing out",
				)
				opts.Timeout = timeout
				flags.BoolVar(&opts.Verbose, "verbose", false, "output stuff as it happens")
				flags.Usage = func() {
					fmt.Printf(`Usage: %s %s

	Parse an Evernote export file and convert data to JSON entities.
	For more info on exporting Evernote data, see:
	%s`,
						_Bin, subcmdName, helpLink)

					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				a.fetchWriteOpts = &opts
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				return interactor.WriteENEXToJSON(
					context.TODO(),
					a.fetchWriteOpts,
				)
			},
		},
		"inspect": &command{
			description: "inspect ENEX file as golang values",
			setup: func(a *arguments) *flag.FlagSet {
				subcmdName := cmdName + " inspect"
				var opts enex.FileOpts
				flags := flag.NewFlagSet(subcmdName, flag.ExitOnError)
				flags.StringVar(
					&opts.Filename,
					"input",
					"",
					"path to evernote export file",
				)
				flags.BoolVar(&opts.PrettyPrint, "pretty", false, "pretty print output")
				flags.Usage = func() {
					fmt.Printf(`Usage: %s %s

	Parse, read an Evernote ENEX file, convert to golang values, output to stdout.
	Could be useful for debugging or inspecting data in development.
	For more info on exporting Evernote data, see:
	%s`,
						_Bin, subcmdName, helpLink)

					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				a.enexExportOpts = &opts
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				return enex.ReadPrintFile(
					context.TODO(),
					a.enexExportOpts,
				)
			},
		},
		"to-sn": &command{
			description: "convert ENEX file to StandardNotes format",
			setup: func(a *arguments) *flag.FlagSet {
				subcmdName := cmdName + " to-sn"
				var opts interactor.ConvertOptions
				flags := flag.NewFlagSet(subcmdName, flag.ExitOnError)
				flags.StringVar(
					&opts.InputFilename,
					"input",
					"",
					"path to evernote export file",
				)
				flags.StringVar(
					&opts.OutputFilename,
					"output",
					"",
					"path to output file",
				)
				flags.Usage = func() {
					fmt.Printf(`Usage: %s %s

	Parse, read an Evernote ENEX file, convert to StandardNotes JSON format.`,
						_Bin, subcmdName)

					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				a.convertOpts = &opts
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				_, err := interactor.ConvertENEXToStandardNotes(
					context.Background(),
					*a.convertOpts,
				)
				return err
			},
		},
	}

	descriptions := describeSubcommands(cmd.subs)
	cmd.flags = flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmd.flags.Usage = func() {
		mainsubName := _Bin + " " + cmdName
		fmt.Printf(`Usage: %s

Description:

	%s handles Evernote export data in the ENEX format.
	For more info on exporting Evernote data, see:
	%s

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`, mainsubName, mainsubName, helpLink, strings.Join(descriptions, "\n\t"),
		)
	}

	return cmd
}("enex")
