package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	fp "path/filepath"
	"strings"
	"time"

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

	newPageCmd = &cobra.Command{
		Use:   "page [title]",
		Short: "Create a new page with specified title",
		Args:  cobra.ExactArgs(1),
		Run:   newPageCmdHandler,
	}

	newPostCmd = &cobra.Command{
		Use:   "post [title]",
		Short: "Create a new post with specified title",
		Args:  cobra.ExactArgs(1),
		Run:   newPostCmdHandler,
	}
)

func init() {
	newSiteCmd.Flags().Bool("force", false, "Init inside non-empty directory")
	newCmd.AddCommand(newSiteCmd, newThemeCmd, newPageCmd, newPostCmd)
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
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please input data for new website")
	fmt.Println()

	cBold.Print("Website title : ")
	tempBytes, _, _ := reader.ReadLine()

	title := string(tempBytes)
	title = strings.TrimSpace(title)
	if title == "" {
		cError.Println("Error: Website title must not empty")
		return
	}

	cBold.Print("Base URL      : ")
	tempBytes, _, _ = reader.ReadLine()

	baseURL := string(tempBytes)
	if _, err = url.ParseRequestURI(baseURL); err != nil {
		cError.Println("Error: Base URL must be an absolute URL path")
		return
	}

	cBold.Print("Website owner : ")
	tempBytes, _, _ = reader.ReadLine()
	owner := string(tempBytes)
	owner = strings.TrimSpace(owner)

	// Create subdirectory
	os.MkdirAll(fp.Join(path, "static"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "theme"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "page"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "post"), os.ModePerm)

	// Write config file
	configPath := fp.Join(path, "config.toml")
	configFile, err := os.Create(configPath)
	if err != nil {
		cError.Println("Error:", err)
		return
	}
	defer configFile.Close()

	err = toml.NewEncoder(configFile).Encode(&Config{
		BaseURL:    baseURL,
		Title:      title,
		Owner:      owner,
		Pagination: 10})
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Finish
	fmt.Println()
	fmt.Print("Congratulations! Your new Spook site is created in ")
	cBold.Println(absPath)
	fmt.Println("Don't forget to check your config file and choose your theme.")
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

	// Make sure valid config file exists in current working dir
	config := Config{}
	_, err = toml.DecodeFile("config.toml", &config)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	if config.BaseURL == "" {
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
	createFile(fp.Join(path, "_base.html"))
	createFile(fp.Join(path, "index.html"))
	createFile(fp.Join(path, "list.html"))
	createFile(fp.Join(path, "page.html"))
	createFile(fp.Join(path, "post.html"))
	createFile(fp.Join(path, "404.html"))

	// Finish
	fmt.Print("Congratulations! Your new theme is created in ")
	cBold.Println(absPath)
}

func newPageCmdHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	title := args[0]

	// Make sure valid config file exists in current working dir
	config := Config{}
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	if config.BaseURL == "" {
		cError.Println("Error: No base URL set in configuration file")
		return
	}

	// Prepare directory name with max length 80 character
	dirPath := ""
	for _, word := range strings.Fields(title) {
		dirPath += strings.ToLower(word) + "-"
		if len(dirPath) >= 80 {
			break
		}
	}

	// Create unique directory name
	dirPath = fp.Join("page", dirPath[:len(dirPath)-1])
	for {
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			dirPath += "-1"
			continue
		} else {
			break
		}
	}

	// Create new directory and index file for the page
	os.MkdirAll(dirPath, os.ModePerm)
	indexFile, err := os.Create(fp.Join(dirPath, "_index.md"))
	if err != nil {
		cError.Println("Error:", err)
		return
	}
	defer indexFile.Close()

	// Write page's metadata
	w := bufio.NewWriter(indexFile)
	fmt.Fprintln(w, "+++")
	toml.NewEncoder(w).Encode(&Page{Title: title})
	fmt.Fprintln(w, "+++")
	w.Flush()

	// Finish
	absPath, err := fp.Abs(dirPath)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	fmt.Print("Congratulations! Your new page is created in ")
	cBold.Println(absPath)
}

func newPostCmdHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	title := args[0]

	// Make sure valid config file exists in current working dir
	config := Config{}
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	if config.BaseURL == "" {
		cError.Println("Error: No base URL set in configuration file")
		return
	}

	// Prepare directory name with max length 100 character
	now := time.Now()
	dirPath := now.Format("2006-01-02-")
	for _, word := range strings.Fields(title) {
		dirPath += strings.ToLower(word) + "-"
		if len(dirPath) >= 100 {
			break
		}
	}

	// Create unique directory name
	dirPath = fp.Join("post", dirPath[:len(dirPath)-1])
	for {
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			dirPath += "-1"
			continue
		} else {
			break
		}
	}

	// Create new directory and index file for the page
	os.MkdirAll(dirPath, os.ModePerm)
	indexFile, err := os.Create(fp.Join(dirPath, "_index.md"))
	if err != nil {
		cError.Println("Error:", err)
		return
	}
	defer indexFile.Close()

	// Write post's metadata
	strNow := now.Format("2006-01-02 15:04:05 -0700")
	metadata := Post{
		Title:     title,
		Excerpt:   "",
		CreatedAt: strNow,
		UpdatedAt: strNow,
		Category:  "",
		Tags:      []string{},
		Author:    config.Owner,
	}

	w := bufio.NewWriter(indexFile)
	fmt.Fprintln(w, "+++")
	toml.NewEncoder(w).Encode(&metadata)
	fmt.Fprintln(w, "+++")
	w.Flush()

	// Finish
	absPath, err := fp.Abs(dirPath)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	fmt.Print("Congratulations! Your new post is created in ")
	cBold.Println(absPath)
}
