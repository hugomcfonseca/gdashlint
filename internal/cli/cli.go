package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/app"
	"github.com/spf13/cobra"
)

// BuildInfo describes build-time metadata injected by release tooling.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// Run is the CLI entrypoint.
func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	root := newRootCommand(stdin, stdout, stderr, build)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "%v\n", err); writeErr != nil {
			return 2
		}
		return 2
	}
	return commandExitCode(root)
}

type exitCodeKey struct{}

func setCommandExitCode(cmd *cobra.Command, code int) {
	cmd.SetContext(context.WithValue(cmd.Context(), exitCodeKey{}, code))
}

func commandExitCode(cmd *cobra.Command) int {
	value := cmd.Context().Value(exitCodeKey{})
	code, ok := value.(int)
	if !ok {
		return 0
	}
	return code
}

func newRootCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer, build BuildInfo) *cobra.Command {
	var showVersion bool

	root := &cobra.Command{
		Use:           "gdashlint",
		Short:         "Lint Grafana dashboard JSON files",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showVersion {
				_, err := fmt.Fprintf(stdout, "gdashlint %s (%s, %s)\n", build.Version, build.Commit, build.Date)
				return err
			}
			return cmd.Help()
		},
	}
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.PersistentFlags().BoolVar(&showVersion, "version", false, "print version information")

	root.AddCommand(newLintCommand(stdin, stdout, stderr), newFixCommand(stdin, stdout, stderr), newRulesCommand(stdout))
	return root
}

func newLintCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *cobra.Command {
	var opts app.Options
	opts.Stdin = stdin
	opts.Stdout = stdout
	opts.Stderr = stderr

	cmd := &cobra.Command{
		Use:          "lint [paths...]",
		Short:        "Lint dashboard JSON files",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Paths = args
			code, err := app.Lint(cmd.Context(), opts)
			setCommandExitCode(cmd.Root(), code)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.ConfigPath, "config", "", "path to config file")
	cmd.Flags().StringVar(&opts.Format, "format", "", "output format: text, json, or github")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "sort mode: severity or file")
	cmd.Flags().StringVar(&opts.FailOn, "fail-on", "", "minimum severity that fails: error, warning, info, or none")
	return cmd
}

func newFixCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *cobra.Command {
	var opts app.Options
	opts.Stdin = stdin
	opts.Stdout = stdout
	opts.Stderr = stderr
	opts.FixMode = "in-place"
	opts.FixSuffix = ".fixed"

	cmd := &cobra.Command{
		Use:          "fix [paths...]",
		Short:        "Apply safe automatic dashboard remediations",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Paths = args
			code, err := app.Fix(cmd.Context(), opts)
			setCommandExitCode(cmd.Root(), code)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.ConfigPath, "config", "", "path to config file")
	cmd.Flags().StringVar(&opts.Format, "format", "", "output format: text, json, or github")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "sort mode: severity or file")
	cmd.Flags().StringVar(&opts.FailOn, "fail-on", "", "minimum severity that fails: error, warning, info, or none")
	cmd.Flags().StringVar(&opts.FixMode, "mode", "in-place", "fix write mode: in-place or copy")
	cmd.Flags().StringVar(&opts.FixSuffix, "suffix", ".fixed", "suffix used before the extension in copy mode")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "show remediations without writing files")
	return cmd
}

func newRulesCommand(stdout io.Writer) *cobra.Command {
	var opts app.Options
	opts.Stdout = stdout

	cmd := &cobra.Command{
		Use:          "rules",
		Short:        "List available rules",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			code, err := app.Rules(opts)
			setCommandExitCode(cmd.Root(), code)
			return err
		},
	}
	cmd.Flags().StringVar(&opts.ConfigPath, "config", "", "path to config file")
	cmd.Flags().StringVar(&opts.Format, "format", "", "output format: text, json, or github")
	return cmd
}
