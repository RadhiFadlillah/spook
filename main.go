package main

import (
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cBold  = color.New(color.Bold)
	cError = color.New(color.FgHiRed)
)

func main() {
	// Create root command
	spookCmd := &cobra.Command{
		Use:   "spook",
		Short: "Simple and minimal static site generator",
	}

	// Execute
	spookCmd.AddCommand(newCmd)
	if err := spookCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
