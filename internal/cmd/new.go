package cmd

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	siteCmd := &cobra.Command{
		Use:   "site [path]",
		Short: "Create a skeleton for new website in specified path",
		Args:  cobra.ExactArgs(1),
		Run:   newSiteHandler,
	}

	themeCmd := &cobra.Command{
		Use:   "theme [name]",
		Short: "Create a skeleton for new theme called [nam] in current directory",
		Args:  cobra.ExactArgs(1),
		Run:   newThemeHandler,
	}

	contentCmd := &cobra.Command{
		Use:   "content [path]",
		Short: "Create a new content in specified path",
		Args:  cobra.ExactArgs(1),
		Run:   newContentHandler,
	}

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new website, theme or content",
	}

	cmd.AddCommand(siteCmd, themeCmd, contentCmd)
	cmd.PersistentFlags().Bool("force", false, "rewrite non-empty directory")
	return cmd
}

func newSiteHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	rootDir := args[0]
	rootDir, _ = fp.Abs(rootDir)
	isForced, _ := cmd.PersistentFlags().GetBool("force")

	// Make sure target directory exists
	os.MkdirAll(rootDir, os.ModePerm)

	// Make sure target dir is empty
	if !dirIsEmpty(rootDir) && !isForced {
		logrus.Fatalf("Directory %s already exists and not empty", rootDir)
	}

	// Get website name and owner from user
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please input data for new website")
	fmt.Println()

	fmt.Print("Website title : ")
	scanner.Scan()
	title := scanner.Text()
	title = strings.TrimSpace(title)

	if title == "" {
		logrus.Fatalln("Website title must not empty")
	}

	fmt.Print("Website owner : ")
	scanner.Scan()
	owner := scanner.Text()
	owner = strings.TrimSpace(owner)

	// Create subdirectory
	os.MkdirAll(fp.Join(rootDir, "content"), os.ModePerm)
	os.MkdirAll(fp.Join(rootDir, "theme"), os.ModePerm)

	// Write config file
	configPath := fp.Join(rootDir, "config.toml")
	configFile, err := os.Create(configPath)
	if err != nil {
		logrus.Fatalln("Failed to create config file:", err)
	}
	defer configFile.Close()

	err = toml.NewEncoder(configFile).Encode(struct {
		Title string
		Owner string
	}{title, owner})

	if err != nil {
		logrus.Fatalln("Failed to write config file:", err)
	}

	// Finish
	fmt.Println()
	fmt.Print("Congratulations! Your new site is created in ")
	fmt.Println(rootDir)
	fmt.Println("Don't forget to check your config file and choose your theme.")
}

func newThemeHandler(cmd *cobra.Command, args []string) {
}

func newContentHandler(cmd *cobra.Command, args []string) {
}
