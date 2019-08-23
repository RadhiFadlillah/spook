package cmd

import (
	"fmt"
	"os"
	fp "path/filepath"

	"github.com/spf13/cobra"
)

func newThemeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "theme [name]",
		Short: "Create a skeleton for new theme",
		Args:  cobra.ExactArgs(1),
		Run:   newThemeHandler,
	}
}

func newThemeHandler(cmd *cobra.Command, args []string) {
	// Read arguments
	name := args[0]
	themeDir := fp.Join("theme", name)
	themeDir, _ = fp.Abs(themeDir)

	// Make sure valid config file exists in current working dir
	_, err := openConfigFile(false)
	if err != nil {
		cError.Println("Failed to open config file:", err)
		return
	}

	// Create new directory for theme
	os.MkdirAll(themeDir, os.ModePerm)

	// Make sure target dir is empty
	if !isEmpty(themeDir) {
		cError.Printf("%s already exists and not empty\n", themeDir)
		return
	}

	// Create directories and files
	os.MkdirAll(fp.Join(themeDir, "res"), os.ModePerm)
	os.MkdirAll(fp.Join(themeDir, "css"), os.ModePerm)
	os.MkdirAll(fp.Join(themeDir, "js"), os.ModePerm)
	createFile(fp.Join(themeDir, "_base.html"))
	createFile(fp.Join(themeDir, "frontpage.html"))
	createFile(fp.Join(themeDir, "list.html"))
	createFile(fp.Join(themeDir, "page.html"))
	createFile(fp.Join(themeDir, "post.html"))
	createFile(fp.Join(themeDir, "404.html"))

	// Finish
	fmt.Print("Congratulations! Your new theme is created in ")
	cBold.Println(themeDir)
}
