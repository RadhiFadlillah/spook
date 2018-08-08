package main

import (
	"fmt"
	"io"
	"net/url"
	"os"
	fp "path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new website or theme",
		Args:  cobra.ExactArgs(1),
	}

	newSiteCmd = &cobra.Command{
		Use:   "site [path]",
		Short: "Create a skeleton for new website and put it inside the provided directory",
		Args:  cobra.ExactArgs(1),
		Run:   newSiteCmdHandler,
	}

	newThemeCmd = &cobra.Command{
		Use:   "theme [name]",
		Short: "Create a skeleton for new theme",
		Args:  cobra.ExactArgs(1),
		Run:   newThemeCmdHandler,
	}
)

func init() {
	newSiteCmd.Flags().Bool("force", false, "Init inside non-empty directory")
	newCmd.AddCommand(newSiteCmd, newThemeCmd)
}

func newSiteCmdHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	isForced, _ := cmd.Flags().GetBool("force")

	path := args[0]
	absPath, err := fp.Abs(path)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Make sure target directory exists
	os.MkdirAll(path, os.ModePerm)

	// Make sure target dir is empty
	targetDir, err := os.Open(path)
	if err != nil {
		cError.Println("Error:", err)
		return
	}
	defer targetDir.Close()

	_, err = targetDir.Readdirnames(1)
	if err != io.EOF && !isForced {
		cError.Printf("Error: %s already exists and not empty\n", absPath)
		return
	}

	// Get website name and base url from user
	var title, baseURL string
	fmt.Println("Please input data for new website")
	fmt.Println()

	cBold.Print("Website title : ")
	fmt.Scanln(&title)

	title = strings.TrimSpace(title)
	if title == "" {
		cError.Println("Error: Website title must not empty")
		return
	}

	cBold.Print("Base URL      : ")
	fmt.Scanln(&baseURL)
	if _, err = url.ParseRequestURI(baseURL); err != nil {
		cError.Println("Error: Base URL must be an absolute URL path")
		return
	}
	fmt.Println()

	// Create subdirectory
	os.MkdirAll(fp.Join(path, "static"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "theme"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "page"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "post"), os.ModePerm)

	// Write config file
	configPath := fp.Join(path, "config.toml")
	configFile, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		cError.Println("Error:", err)
		return
	}
	defer configFile.Close()

	err = toml.NewEncoder(configFile).Encode(&map[string]string{
		"baseURL": baseURL,
		"title":   title})
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Finish
	fmt.Print("Congratulations! Your new Spook site is created in ")
	cBold.Println(absPath)
}

func newThemeCmdHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	name := args[0]
	path := fp.Join("theme", name)
	absPath, err := fp.Abs(path)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Open config file in current working dir
	mapConfig := map[string]string{}
	_, err = toml.DecodeFile("config.toml", &mapConfig)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	if _, exist := mapConfig["baseURL"]; !exist {
		cError.Println("Error: No base URL set in configuration file")
		return
	}

	// Create new directory for theme
	os.MkdirAll(path, os.ModePerm)

	// Make sure target dir is empty
	targetDir, err := os.Open(path)
	if err != nil {
		cError.Println("Error:", err)
		return
	}
	defer targetDir.Close()

	_, err = targetDir.Readdirnames(1)
	if err != io.EOF {
		cError.Printf("Error: %s already exists and not empty\n", absPath)
		return
	}

	// Create directories and files
	os.MkdirAll(fp.Join(path, "res"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "css"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "js"), os.ModePerm)
	createFile(fp.Join(path, "index.html"))
	createFile(fp.Join(path, "page.html"))
	createFile(fp.Join(path, "post.html"))
	createFile(fp.Join(path, "404.html"))
}
