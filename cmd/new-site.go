package cmd

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-spook/spook/model"
	"github.com/spf13/cobra"
)

func newSiteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "site [path]",
		Short: "Create a skeleton for new website and put it inside the provided directory",
		Args:  cobra.ExactArgs(1),
		Run:   newSiteHandler,
	}

	cmd.Flags().Bool("force", false, "force init inside non-empty directory")

	return cmd
}

func newSiteHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	rootDir := args[0]
	rootDir, _ = fp.Abs(rootDir)
	isForced, _ := cmd.Flags().GetBool("force")

	// Make sure target directory exists
	os.MkdirAll(rootDir, os.ModePerm)

	// Make sure target dir is empty
	if !isEmpty(rootDir) && !isForced {
		cError.Printf("Directory %s already exists and not empty\n", rootDir)
		return
	}

	// Get website name and base url from user
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input data for new website")
	fmt.Println()

	cBold.Print("Website title : ")
	scanner.Scan()

	title := scanner.Text()
	title = strings.TrimSpace(title)
	if title == "" {
		cError.Println("Website title must not empty")
		return
	}

	cBold.Print("Website owner : ")
	scanner.Scan()

	owner := scanner.Text()
	owner = strings.TrimSpace(owner)

	// Create subdirectory
	os.MkdirAll(fp.Join(rootDir, "static"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "theme"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "page"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "post"), os.ModePerm)

	// Write config file
	configPath := fp.Join(rootDir, "config.toml")
	configFile, err := os.Create(configPath)
	if err != nil {
		cError.Println("Failed to create config file:", err)
		return
	}
	defer configFile.Close()

	err = toml.NewEncoder(configFile).Encode(&model.Config{
		Title:      title,
		Owner:      owner,
		Pagination: 10})
	if err != nil {
		cError.Println("Failed to write config file:", err)
		return
	}

	// Finish
	fmt.Println()
	fmt.Print("Congratulations! Your new Spook site is created in ")
	cBold.Println(rootDir)
	fmt.Println("Don't forget to check your config file and choose your theme.")
}
