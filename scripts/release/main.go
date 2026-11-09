// Command release renders and publishes source-built Homebrew formulas.
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/the-inconvenience-store/pbj/internal/release"
)

func main() {
	root := newCommand()
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	root := &cobra.Command{Use: "release", Short: "Prepare source-built Homebrew releases", Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true, CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true}}
	var validationTag string
	validate := &cobra.Command{Use: "validate", Short: "Validate a stable vMAJOR.MINOR.PATCH release tag", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error { return release.ValidateTag(validationTag) }}
	validate.Flags().StringVar(&validationTag, "tag", "", "Stable release tag")
	root.AddCommand(validate)
	var tag, source, output string
	formula := &cobra.Command{Use: "formula", Short: "Render a formula using the actual source archive checksum", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		if output == "" {
			return errors.New("--output is required")
		}
		prepared, err := release.Prepare(tag, source)
		if err != nil {
			return err
		}
		const formulaPermissions = 0o600
		return os.WriteFile(output, prepared.Formula(), formulaPermissions)
	}}
	formula.Flags().StringVar(&tag, "tag", "", "Stable release tag")
	formula.Flags().StringVar(&source, "source", "", "Tagged GoReleaser source archive")
	formula.Flags().StringVar(&output, "output", "", "Destination formula file")
	root.AddCommand(formula)
	var publishTag, publishSource string
	publish := &cobra.Command{Use: "publish", Short: "Publish the matching source-built formula to the tap", Args: cobra.NoArgs, RunE: func(command *cobra.Command, _ []string) error {
		prepared, err := release.Prepare(publishTag, publishSource)
		if err != nil {
			return err
		}
		const requestTimeout = 30 * time.Second
		publisher := release.Publisher{Client: &http.Client{Timeout: requestTimeout}, Endpoint: release.TapEndpoint, Token: os.Getenv("HOMEBREW_TAP_TOKEN")}
		changed, err := publisher.Publish(command.Context(), prepared)
		if err != nil {
			return err
		}
		if changed {
			_, err = fmt.Fprintln(command.OutOrStdout(), "Published source-built pbj formula")
		} else {
			_, err = fmt.Fprintln(command.OutOrStdout(), "Tap formula already matches this release")
		}
		return err
	}}
	publish.Flags().StringVar(&publishTag, "tag", "", "Stable release tag")
	publish.Flags().StringVar(&publishSource, "source", "", "Published GoReleaser source archive")
	root.AddCommand(publish)
	return root
}
