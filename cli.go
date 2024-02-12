package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

// ErrFailure can be returned from a command to trigger a non-zero exit code with no extra message.
var ErrFailure = errors.New("failure")

// PrintErrors is intentionally excluded because it is not applied to internal errors
// (such as default value on a boolean flag).
// We want to print these internal errors, but don't want to print user errors twice.
var flagsOptions flags.Options = flags.HelpFlag | flags.PassDoubleDash

// Initer is an optional interface for commands that require global flags.
// The command can read the global values and store them to be later used in Execute.
type Initer[T any] interface {
	Init(app *T) error
}

// Validator is an optional interface for commands that require additional validation of their flags.
type Validator interface {
	Validate() error
}

// Commander is a required interface for CLI commands.
type Commander interface {
	Execute(args []string) error
}

func ParseExecute[T any]() *T {
	var app T

	parser := flags.NewParser(&app, flagsOptions)
	parser.CommandHandler = func(command flags.Commander, args []string) error {
		if command == nil {
			return nil
		}

		if validator, ok := command.(Validator); ok {
			if err := validator.Validate(); err != nil {
				handleError(err)
			}
		}

		if initer, ok := command.(Initer[T]); ok {
			if err := initer.Init(&app); err != nil {
				handleError(err)
			}
		}

		if err := command.Execute(args); err != nil {
			handleError(err)
		}

		return nil
	}

	if _, err := parser.Parse(); err != nil {
		handleError(err)
	}

	return &app
}

func Parse[T any]() *T {
	var app T
	parser := flags.NewParser(&app, flagsOptions)
	if _, err := parser.Parse(); err != nil {
		handleError(err)
	}
	return &app
}

func handleError(err error) {
	if err == nil {
		return
	}

	if errors.Is(err, ErrFailure) {
		os.Exit(1)
	}

	var flagErr *flags.Error
	if errors.As(err, &flagErr) {
		if flagErr.Type == flags.ErrHelp {
			fmt.Println(err)
			os.Exit(0)
		}
		// Intentionally fall through to print the flags error
	}

	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
