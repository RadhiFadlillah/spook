package cmd

import (
	"fmt"
	"os"
	fp "path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-spook/spook/model"
	"github.com/spf13/cobra"
)

func newPostCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "post [title]",
		Short: "Create a new post with specified title",
		Args:  cobra.ExactArgs(1),
		Run:   newPostHandler,
	}
}

func newPostHandler(cmd *cobra.Command, args []string) {
	// Make sure valid config file exists in current working dir
	config, err := openConfigFile(false)
	if err != nil {
		cError.Println("Failed to open config file:", err)
		return
	}

	// Get current time
	now := time.Now()
	date := now.Format("2006-01-02")
	dateTime := now.Format("2006-01-02 15:04:05 -0700")

	// Create unique directory name with max length 90 character
	title := args[0]
	postName := createDirName(title, "post", 90)
	postName = date + "-" + postName

	// Create post dir
	postDir := fp.Join("post", postName)
	postDir, _ = fp.Abs(postDir)
	os.MkdirAll(postDir, os.ModePerm)

	// Create index file for the post
	indexPath := fp.Join(postDir, "_index.md")
	indexFile, err := os.Create(indexPath)
	if err != nil {
		cError.Println("Failed to create index file:", err)
		return
	}
	defer indexFile.Close()

	// Write post's metadata
	metadata := model.Post{
		Title:     title,
		Excerpt:   "",
		CreatedAt: dateTime,
		UpdatedAt: dateTime,
		Category:  "",
		Tags:      []string{},
		Author:    config.Owner,
	}

	fmt.Fprintln(indexFile, "+++")
	toml.NewEncoder(indexFile).Encode(&metadata)
	fmt.Fprintln(indexFile, "+++")
	indexFile.Sync()

	// Finish
	fmt.Print("Congratulations! Your new post is created in ")
	cBold.Println(postDir)
}
