package cmd

import (
	"fmt"
	"os"
	fp "path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/go-spook/spook/model"
	"github.com/spf13/cobra"
)

func newPageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "page [title]",
		Short: "Create a new page with specified title",
		Args:  cobra.ExactArgs(1),
		Run:   newPageHandler,
	}
}

func newPageHandler(cmd *cobra.Command, args []string) {
	// Make sure valid config file exists in current working dir
	_, err := openConfigFile(false)
	if err != nil {
		cError.Println("Failed to open config file:", err)
		return
	}

	// Create unique directory name with max length 80 character
	title := args[0]
	pageName := createDirName(title, "page", 80)

	// Create page dir
	pageDir := fp.Join("page", pageName)
	pageDir, _ = fp.Abs(pageDir)
	os.MkdirAll(pageDir, os.ModePerm)

	// Create index file for the page
	indexPath := fp.Join(pageDir, "_index.md")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		cError.Println("Failed to create index file:", err)
		return
	}
	defer indexFile.Close()

	// Write page's metadata
	metadata := model.Page{
		Title: title,
	}

	fmt.Fprintln(indexFile, "+++")
	toml.NewEncoder(indexFile).Encode(&metadata)
	fmt.Fprintln(indexFile, "+++")
	indexFile.Sync()

	// Finish
	fmt.Print("Congratulations! Your new page is created in ")
	cBold.Println(pageDir)
}
