// Package cmd wraps up all command line interface operations.
package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
	"github.com/rafaelespinoza/notexfr/internal/repo/enex"
)

// An arguments captures inputs, options for the main command and subcommands.
type arguments struct {
	positionalArgs []string
	production     bool
	envfile        string
	fetchWriteOpts *interactor.FetchWriteParams
	enexExportOpts *enex.FileOpts
	backfillOpts   *interactor.BackfillParams
	convertOpts    *interactor.ConvertParams
}

var (
	// _Args is a shared top-level arguments value.
	_Args arguments
	// _Bin is the name of the binary file.
	_Bin = os.Args[0]
	// _MainCommand is the parent command for subcommands and their children.
	_MainCommand *delegator
)

// Init should be invoked in the init func in package main.
func Init() {
	_MainCommand = &delegator{
		description: "main command for " + _Bin,
		subs: map[string]directive{
			"backfill": _Backfill,
			"convert":  _Convert,
			"edam":     _Edam,
			"enex":     _Enex,
			"version":  _Version,
		},
	}

	flag.Usage = func() {
		descriptions := describeSubcommands(_MainCommand.subs)
		fmt.Fprintf(flag.CommandLine.Output(), `Usage:
	%s [flags] subcommand [subflags]

Description:

	%s is a tool for converting note data to other service formats.

	Currently supported services:
	- Evernote
	- StandardNotes
	`,
			_Bin, _Bin)

		fmt.Fprintf(flag.CommandLine.Output(), `
Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v

Examples:

	%s [subcommand] -h
`,
			strings.Join(descriptions, "\n\t"), _Bin)
	}
}

// Run should be invoked in the main func in package main.
func Run(ctx context.Context) (err error) {
	flag.Parse()
	_Args.positionalArgs = flag.Args()
	var deleg directive
	err = _MainCommand.perform(ctx, &_Args)
	if _MainCommand.selected == nil {
		// either asked for help or asked for unknown command.
		flag.Usage()
	} else {
		deleg = _MainCommand.selected
	}
	if err != nil {
		return
	}

	if _, ok := deleg.(*command); ok {
		return deleg.perform(ctx, &_Args)
	}

	// a panic is possible here, but all direct children of the main command
	// are delegators so it's not designed to happen.
	topic := deleg.(*delegator)
	if err = topic.perform(ctx, &_Args); err != nil {
		topic.flags.Usage()
		return
	}

	switch subcmd := topic.selected.(type) {
	case *command:
		err = subcmd.perform(ctx, &_Args)
	case *delegator:
		err = fmt.Errorf("too much delegation, selected should be a %T", &command{})
	default:
		err = fmt.Errorf("unhandled type %T", subcmd)
	}
	return
}

// directive is an abstraction for a parent or child command. A parent would
// delegate to a subcommand, while a subcommand does the actual task.
type directive interface {
	// summary provides a short, one-line description.
	summary() string
	// perform should either choose a subcommand or do a task.
	perform(ctx context.Context, a *arguments) error
}

// A delegator is a parent to a set of commands. Its sole purpose is to direct
// traffic to a selected command. It can also collect common flag inputs to pass
// on to subcommands.
type delegator struct {
	// description should provide a short summary.
	description string
	// flags collect and share inputs to its sub directives.
	flags *flag.FlagSet
	// selected is the chosen transfer point of control.
	selected directive
	// subs associates a name with a link to another directive. NOTE: one does
	// not simply create too many layers of delegators.
	subs map[string]directive
}

func (d *delegator) summary() string { return d.description }

// perform chooses a subcommand.
func (d *delegator) perform(ctx context.Context, a *arguments) error {
	if len(a.positionalArgs) < 1 {
		return flag.ErrHelp
	}
	var err error
	switch a.positionalArgs[0] {
	case "-h", "-help", "--help", "help":
		err = flag.ErrHelp
	default:
		if cmd, ok := d.subs[a.positionalArgs[0]]; !ok {
			err = fmt.Errorf("unknown command %q", a.positionalArgs[0])
		} else {
			d.selected = cmd
		}
	}
	if err != nil {
		return err
	}

	switch selected := d.selected.(type) {
	case *command:
		err = selected.setup(a).Parse(a.positionalArgs[1:])
	case *delegator:
		a.positionalArgs = a.positionalArgs[1:] // I, also like to live dangerously
	default:
		err = fmt.Errorf("unsupported value of type %T", selected)
	}
	return err
}

// A command performs a task.
type command struct {
	// description should provide a short summary.
	description string
	// setup should prepare Args for interpretation by using the pointer to Args
	// with the returned flag set.
	setup func(a *arguments) *flag.FlagSet
	// run is a wrapper function that selects the necessary command line inputs,
	// executes the command and returns any errors.
	run func(ctx context.Context, a *arguments) error
}

func (c *command) summary() string                                 { return c.description }
func (c *command) perform(ctx context.Context, a *arguments) error { return c.run(ctx, a) }

func describeSubcommands(subcmds map[string]directive) []string {
	descriptions := make([]string, 0)
	for name, subcmd := range subcmds {
		descriptions = append(
			descriptions,
			fmt.Sprintf("%-20s\t%-40s", name, subcmd.summary()),
		)
	}
	sort.Strings(descriptions)
	return descriptions
}
