package cmd

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-spook/spook/model"
	"github.com/spf13/cobra"
)

var (
	cmdNew = &cobra.Command{
		Use:   "new",
		Short: "Create a new website, theme or content",
	}

	cmdNewSite = &cobra.Command{
		Use:   "site [path]",
		Short: "Create a skeleton for new website and put it inside the provided directory",
		Args:  cobra.ExactArgs(1),
		Run:   cmdNewSiteHandler,
	}

	cmdNewTheme = &cobra.Command{
		Use:   "theme [name]",
		Short: "Create a skeleton for new theme",
		Args:  cobra.ExactArgs(1),
		Run:   cmdNewThemeHandler,
	}

	cmdNewPage = &cobra.Command{
		Use:   "page [title]",
		Short: "Create a new page with specified title",
		Args:  cobra.ExactArgs(1),
		Run:   cmdNewPageHandler,
	}

	cmdNewPost = &cobra.Command{
		Use:   "post [title]",
		Short: "Create a new post with specified title",
		Args:  cobra.ExactArgs(1),
		Run:   cmdNewPostHandler,
	}
)

func init() {
	cmdNewSite.Flags().Bool("force", false, "Init inside non-empty directory")
	cmdNew.AddCommand(cmdNewSite, cmdNewTheme, cmdNewPage, cmdNewPost)
}

func cmdNewSiteHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	path := args[0]
	absPath, _ := fp.Abs(path)
	isForced, _ := cmd.Flags().GetBool("force")

	// Make sure target directory exists
	os.MkdirAll(path, os.ModePerm)

	// Make sure target dir is empty
	if !isEmpty(path) && !isForced {
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

	err = toml.NewEncoder(configFile).Encode(&model.Config{
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

func cmdNewThemeHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	name := args[0]
	path := fp.Join("theme", name)
	absPath, _ := fp.Abs(path)

	// Make sure valid config file exists in current working dir
	_, err := openConfigFile(false)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Create new directory for theme
	os.MkdirAll(path, os.ModePerm)

	// Make sure target dir is empty
	if !isEmpty(path) {
		cError.Printf("Error: %s already exists and not empty\n", absPath)
		return
	}

	// Create directories and files
	os.MkdirAll(fp.Join(path, "res"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "css"), os.ModePerm)
	os.MkdirAll(fp.Join(path, "js"), os.ModePerm)
	createFile(fp.Join(path, "_base.html"))
	createFile(fp.Join(path, "frontpage.html"))
	createFile(fp.Join(path, "list.html"))
	createFile(fp.Join(path, "page.html"))
	createFile(fp.Join(path, "post.html"))
	createFile(fp.Join(path, "404.html"))

	// Finish
	fmt.Print("Congratulations! Your new theme is created in ")
	cBold.Println(absPath)
}

func cmdNewPageHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	title := args[0]

	// Make sure valid config file exists in current working dir
	_, err := openConfigFile(false)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Create unique directory name with max length 80 character
	dirPath := createDirName(title, "page", 80)

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
	toml.NewEncoder(w).Encode(&model.Page{Title: title})
	fmt.Fprintln(w, "+++")
	w.Flush()

	// Finish
	absPath, _ := fp.Abs(dirPath)
	fmt.Print("Congratulations! Your new page is created in ")
	cBold.Println(absPath)
}

func cmdNewPostHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	title := args[0]

	// Make sure valid config file exists in current working dir
	config, err := openConfigFile(false)
	if err != nil {
		cError.Println("Error:", err)
		return
	}

	// Create unique directory name with max length 100 character
	now := time.Now()
	dirName := now.Format("2006-01-02-") + title
	dirPath := createDirName(dirName, "post", 100)

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
	metadata := model.Post{
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
	absPath, _ := fp.Abs(dirPath)
	fmt.Print("Congratulations! Your new post is created in ")
	cBold.Println(absPath)
}
