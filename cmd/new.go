package cmd

import (
	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new website, theme or content",
	}

	cmd.AddCommand(newSiteCmd(), newThemeCmd(), newPageCmd(), newPostCmd())
	return cmd
}
