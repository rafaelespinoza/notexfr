// Package cmd wraps up all command line interface operations.
package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/rafaelespinoza/notexfr/internal/log"
	"github.com/spf13/cobra"
)

const (
	mainName    = "notexfr"
	dataGroupID = "data_group"
)

// Root abstracts a top-level command from package main.
type Root interface {
	// ExecuteContext is the entry point.
	ExecuteContext(ctx context.Context) error
}

// New establishes the root comand and its subcommands.
func New() Root {
	out := cobra.Command{
		Use:   mainName,
		Short: "main command for " + mainName,
		Long: `A tool for converting note data to other service formats.

Currently supported services
- Evernote
- StandardNotes`,
	}
	rootFlags := out.PersistentFlags()
	rootFlags.BoolP("quiet", "q", false, "if true, then all logging is effectively off")
	rootFlags.StringP("log-level", "", validLoggingLevels[len(validLoggingLevels)-1].String(), fmt.Sprintf("minimum severity for which to log events, should be one of %q", validLoggingLevels))
	rootFlags.StringP("log-format", "", validLoggingFormats[len(validLoggingFormats)-1], fmt.Sprintf("output format for logs, should be one of %q", validLoggingFormats))
	out.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		flags := c.Root().PersistentFlags() // find root so that child commands may use this functionality
		loggingOff, err := flags.GetBool("quiet")
		if err != nil {
			return err
		}
		logLevel, err := flags.GetString("log-level")
		if err != nil {
			return err
		}
		logFormat, err := flags.GetString("log-format")
		if err != nil {
			return err
		}

		handler, err := newLogHandler(os.Stderr, loggingOff, logLevel, logFormat)
		if err != nil {
			return err
		}
		log.Init(handler)
		return nil
	}

	out.AddCommand(
		makeBackfill("backfill"),
		makeConvert("convert"),
		makeEdam("edam"),
		makeEnex("enex"),
		makeVersion("version"),
	)
	out.AddGroup(
		&cobra.Group{ID: dataGroupID, Title: "Data Commands:"},
	)
	return &out
}

var (
	validLoggingLevels  = []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	validLoggingFormats = []string{"json", "text"}
)

func newLogHandler(w io.Writer, loggingOff bool, logLevel, logFormat string) (slog.Handler, error) {
	if loggingOff {
		return nil, nil
	}

	lvl := slog.LevelDebug - 1 // sentinel value to help recognize invalid input
	for _, validLevel := range validLoggingLevels {
		if strings.ToUpper(logLevel) == validLevel.String() {
			lvl = validLevel
			break
		}
	}
	if lvl < slog.LevelDebug {
		return nil, fmt.Errorf("invalid log level %q; should be one of %q", logLevel, validLoggingLevels)
	}

	replaceAttrs := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}

	var h slog.Handler
	switch strings.ToLower(logFormat) {
	case "json":
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl, ReplaceAttr: replaceAttrs})
	case "text":
		h = slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl, ReplaceAttr: replaceAttrs})
	default:
		return nil, fmt.Errorf("invalid log format, should be one of %q", validLoggingFormats)
	}

	return h, nil
}
