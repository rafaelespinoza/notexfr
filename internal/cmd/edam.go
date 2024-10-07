package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/rafaelespinoza/notexfr/internal/interactor"
	"github.com/rafaelespinoza/notexfr/internal/repo/edam"
)

func makeEdam(cmdName string) *cobra.Command {
	cmd := cobra.Command{
		Use:     cmdName,
		GroupID: dataGroupID,
		Short:   "handle Evernote data from the EDAM API",
		Long: `Interacts with Evernote via its API.

Handles fetching resources from the Evernote API (EDAM) and writes the
results to local JSON files. A developer token is required to access your
Evernote account.

See the developer guide to get started: https://dev.evernote.com/doc/

Specify account credentials in a file with the --envfile flag.
Use the --production flag to use production credentials, otherwise it will
default to using sandbox credentials.`,
	}

	makeEnv := cobra.Command{
		Use:   "make-env",
		Short: "init an env var file unless it already exists",
		Long:  `Create an environment variable file to store Evernote sandbox and production credentials`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			envfile, err := flags.GetString("envfile")
			if err != nil {
				return err
			}
			return interactor.MakeEDAMEnvFile(envfile)
		},
	}
	{
		makeEnv.Flags().StringP("envfile", "e", "", "path to env var file")
	}

	notebooks := cobra.Command{
		Use:   "notebooks",
		Short: "fetch Notebooks and write data to JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := newEDAMCtx(cmd)
			if err != nil {
				return err
			}
			opts, err := buildEDAMFetchWriteParams(cmd)
			if err != nil {
				return err
			}
			return interactor.FetchWriteNotebooks(ctx, &opts)
		},
	}
	setupEDAMSubcmd(&notebooks)

	notes := cobra.Command{
		Use:   "notes",
		Short: "fetch Notes and write data to JSON file",
	}
	setupEDAMSubcmd(&notes)
	{
		notesFlags := notes.Flags()
		notesFlags.Int32P("lo-index", "L", 0, "start index for paginating notes")
		notesFlags.Int32P("hi-index", "H", -1, "end index for paginating notes, if negative go until there are no more")
		notesFlags.Int32P("page-size", "S", 100, "number of results to fetch at once")

		notes.RunE = func(cmd *cobra.Command, args []string) error {
			ctx, err := newEDAMCtx(cmd)
			if err != nil {
				return err
			}

			opts, err := buildEDAMFetchWriteParams(cmd)
			if err != nil {
				return err
			}
			flags := cmd.Flags()
			var rpq edam.NotesRemoteQueryParams
			rpq.LoIndex, err = flags.GetInt32("lo-index")
			if err != nil {
				return err
			}
			rpq.HiIndex, err = flags.GetInt32("hi-index")
			if err != nil {
				return err
			}
			rpq.PageSize, err = flags.GetInt32("page-size")
			if err != nil {
				return err
			}
			opts.NotesQueryParams = &rpq
			return interactor.FetchWriteNotes(ctx, &opts)
		}
	}

	tags := cobra.Command{
		Use:   "tags",
		Short: "fetch Tags and write data to JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := newEDAMCtx(cmd)
			if err != nil {
				return err
			}

			opts, err := buildEDAMFetchWriteParams(cmd)
			if err != nil {
				return err
			}

			return interactor.FetchWriteTags(ctx, &opts)
		},
	}
	setupEDAMSubcmd(&tags)

	cmd.AddCommand(&makeEnv, &notebooks, &notes, &tags)
	return &cmd
}

func setupEDAMSubcmd(cmd *cobra.Command) {
	cmd.Long = fmt.Sprintf(`Fetch %s from your Evernote account and write them to JSON files.
Use your sandbox account by default. To use your production account, pass
the -production flag.
Specify account credentials with the -envfile flag.`, cmd.Name())

	flags := cmd.Flags()
	flags.StringP("output", "o", "", "path to write data as JSON")
	flags.DurationP("timeout", "t", time.Duration(120)*time.Second, "how long to wait before timing out")
	flags.BoolP("production", "p", false, "use production evernote account")
	flags.StringP("envfile", "e", "", "path to to env var file")
}

func newEDAMCtx(cmd *cobra.Command) (out context.Context, err error) {
	flags := cmd.Flags()

	envfile, err := flags.GetString("envfile")
	if err != nil {
		err = fmt.Errorf("failed to build EDAM context: %w", err)
		return
	}

	var serviceEnvironment edam.EvernoteService
	prod, err := flags.GetBool("production")
	if err != nil {
		err = fmt.Errorf("failed to build EDAM context: %w", err)
		return
	}
	if prod {
		serviceEnvironment = edam.EvernoteProductionService
	}

	val := edam.CredentialsConfig{EnvFilename: envfile, ServiceEnv: serviceEnvironment}
	out = context.WithValue(cmd.Context(), edam.EvernoteServiceKey, val)
	return
}

func buildEDAMFetchWriteParams(cmd *cobra.Command) (out interactor.FetchWriteParams, err error) {
	flags := cmd.Flags()
	out.OutputFilename, err = flags.GetString("output")
	if err != nil {
		return
	}
	out.Timeout, err = flags.GetDuration("timeout")
	if err != nil {
		return
	}
	return
}
