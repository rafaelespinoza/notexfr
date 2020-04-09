package cmd

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
	"github.com/rafaelespinoza/notexfr/internal/repo/edam"
)

var _Edam = func(cmdName string) *delegator {
	cmd := &delegator{description: "handle Evernote data from the EDAM API"}

	// Define this field before defining the Usage function so the subcommand
	// descriptions are present in the help message.
	cmd.subs = map[string]directive{
		"make-env": &command{
			description: "init an env var file unless it already exists",
			setup: func(a *arguments) *flag.FlagSet {
				const name = "make-env"
				flags := flag.NewFlagSet(name, flag.ExitOnError)
				flags.StringVar(
					&a.envfile,
					"envfile",
					"",
					"path to env var file",
				)
				flags.Usage = func() {
					fmt.Printf(`Usage: %s edam %s

	Create an environment variable file to store Evernote sandbox and production
	credentials.`,
						_Bin, name)
					fmt.Printf("\n\nFlags:\n\n")
					flags.PrintDefaults()
				}
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				return interactor.MakeEDAMEnvFile(a.envfile)
			},
		},
		"notebooks": &command{
			description: "fetch Notebooks and write data to JSON file",
			setup: func(a *arguments) *flag.FlagSet {
				return setupFetchWriteCommands(a, "notebooks")
			},
			run: func(ctx context.Context, a *arguments) error {
				return interactor.FetchWriteNotebooks(
					newEdamContext(ctx, a.envfile, a.production),
					a.fetchWriteOpts,
				)
			},
		},
		"notes": &command{
			description: "fetch Notes and write data to JSON file",
			setup: func(a *arguments) *flag.FlagSet {
				flags := setupFetchWriteCommands(a, "notes")
				var listParams edam.NotesRemoteQueryParams
				flags.IntVar(
					&listParams.LoIndex,
					"lo-index",
					0,
					"start index for paginating notes",
				)
				flags.IntVar(
					&listParams.HiIndex,
					"hi-index",
					-1,
					"end index for paginating notes. if negative, go until there are no more.",
				)
				flags.IntVar(
					&listParams.PageSize,
					"page-size",
					100,
					"number of results to fetch at once",
				)
				a.fetchWriteOpts.NotesQueryParams = &listParams
				return flags
			},
			run: func(ctx context.Context, a *arguments) error {
				a.fetchWriteOpts.NotesQueryParams.Verbose = a.fetchWriteOpts.Verbose
				return interactor.FetchWriteNotes(
					newEdamContext(ctx, a.envfile, a.production),
					a.fetchWriteOpts,
				)
			},
		},
		"tags": &command{
			description: "fetch Tags and write data to JSON file",
			setup: func(a *arguments) *flag.FlagSet {
				return setupFetchWriteCommands(a, "tags")
			},
			run: func(ctx context.Context, a *arguments) error {
				return interactor.FetchWriteTags(
					newEdamContext(ctx, a.envfile, a.production),
					a.fetchWriteOpts,
				)
			},
		},
	}

	descriptions := describeSubcommands(cmd.subs)
	cmd.flags = flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmd.flags.Usage = func() {
		mainsubName := _Bin + " " + cmdName
		fmt.Printf(`Usage: %s

Description:

	%s interacts with Evernote.
	It handles fetching resources from the Evernote API (EDAM) and writes the
	results to local JSON files. A developer token is required to access your
	Evernote account.

	See the developer guide to get started:
	%s

	Specify account credentials in a file with the -envfile flag.
	Use the -production flag to use production credentials, otherwise it will
	default to using sandbox credentials.

Subcommands:

	These will have their own set of flags. Put them after the subcommand.

	%v
`, mainsubName, mainsubName, "https://dev.evernote.com/doc/", strings.Join(descriptions, "\n\t"),
		)
	}
	return cmd
}("edam")

func setupFetchWriteCommands(a *arguments, name string) *flag.FlagSet {
	var opts interactor.FetchWriteParams
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	flags.StringVar(
		&opts.OutputFilename,
		"output",
		"",
		"path to write data as JSON",
	)
	flags.DurationVar(
		&opts.Timeout,
		"timeout",
		time.Duration(120)*time.Second,
		"how long to wait before timing out",
	)
	flags.BoolVar(&opts.Verbose, "verbose", false, "output stuff as it happens")
	flags.BoolVar(&_Args.production, "production", false, "use production evernote account")
	flags.StringVar(&_Args.envfile, "envfile", "", "path to to env var file")
	a.fetchWriteOpts = &opts
	flags.Usage = func() {
		fmt.Printf(`Usage: %s edam %s

	Fetch %s from your Evernote account and write them to JSON files.
	Use your sandbox account by default. To use your production account, pass
	the -production flag.
	Specify account credentials with the -envfile flag.`,
			_Bin, name, name)

		fmt.Printf("\n\nFlags:\n\n")
		flags.PrintDefaults()
	}
	return flags
}

func newEdamContext(parentCtx context.Context, envFile string, production bool) (ctx context.Context) {
	var serviceEnvironment edam.EvernoteService
	if production {
		serviceEnvironment = edam.EvernoteProductionService
	}
	ctx = context.WithValue(parentCtx, edam.EvernoteServiceKey, edam.CredentialsConfig{
		EnvFilename: envFile,
		ServiceEnv:  serviceEnvironment,
	})
	return
}
