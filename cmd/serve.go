package cmd

import (
	"os"

	"github.com/go-spook/spook/webserver"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run a webserver that serves the site",
		Long: "Run a simple and performant web server which serves the site. " +
			"Server will avoid writing the rendered and served content to disk, preferring to store it in memory. " +
			"If --port flag is not used, it will use port 8080 by default.",
		Aliases: []string{"serve"},
		Args:    cobra.NoArgs,
		Run:     serveHandler,
	}

	cmd.Flags().IntP("port", "p", 8080, "Port that used by webserver")

	return cmd
}

func serveHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	port, _ := cmd.Flags().GetInt("port")

	// Get working dir
	rootDir, err := os.Getwd()
	if err != nil {
		cError.Println("Failed to get working dir:", err)
		return
	}

	// Make sure valid config file exists in current working dir
	config, err := openConfigFile(true)
	if err != nil {
		cError.Println("Failed to open config file:", err)
		return
	}

	// Start server
	logrus.Printf("Serve spook in :%d\n", port)
	webserver.Start(rootDir, config, port)
}
