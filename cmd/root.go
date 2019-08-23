package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cBold  = color.New(color.Bold)
	cError = color.New(color.FgHiRed)
)

// SpookCmd creates new command for spook
func SpookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spook",
		Short: "Simple, minimalist and opinionated static site generator",
	}

	cmd.AddCommand(newCmd(), serveCmd(), buildCmd())
	return cmd
}
