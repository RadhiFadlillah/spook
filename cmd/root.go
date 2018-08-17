package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cBold  = color.New(color.Bold)
	cError = color.New(color.FgHiRed)
)

// NewSpookCmd creates new command for spook
func NewSpookCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "spook",
		Short: "Simple, minimalist and opinionated static site generator",
	}

	rootCmd.AddCommand(cmdNew, cmdServer, cmdBuild)
	return rootCmd
}
