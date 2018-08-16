package renderer

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"path"
	fp "path/filepath"
	"strings"

	"github.com/RadhiFadlillah/spook/model"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// ListType is the type of list that will be rendered.
type ListType int

const (
	// DEFAULT means the list is default that shows all posts.
	DEFAULT ListType = iota
	// CATEGORY means the list is list that only shows posts with specified category.
	CATEGORY
	// TAG means the list is list that only shows posts with specified tags.
	TAG
)

// Renderer is used to render static HTML file
type Renderer struct {
	Config     model.Config
	Pages      []model.Page
	Posts      []model.Post
	Tags       []model.Group
	Categories []model.Group
}

var funcsMap = template.FuncMap{
	"add":           add,
	"formatTime":    formatTime,
	"limitSentence": limitSentence,
}

// RenderFrontPage renders front page of the site.
// If exists, it will use the front page template.
// If not, it will fallback to using list template.
func (rd Renderer) RenderFrontPage(pageNumber int, dst io.Writer) error {
	// Make sure config file is valid
	err := rd.validateConfig()
	if err != nil {
		return err
	}

	// Prepare templates
	themeDir := rd.Config.Theme
	tplList := fp.Join(themeDir, "list.html")
	tplFrontPage := fp.Join(themeDir, "frontpage.html")

	templates, err := rd.getBaseTemplates()
	if err != nil {
		return err
	}

	activeTemplate := ""
	if fileExists(tplFrontPage) {
		activeTemplate = "frontpage.html"
		templates = append(templates, tplFrontPage)
	} else if fileExists(tplList) {
		activeTemplate = "list.html"
		templates = append(templates, tplList)
	} else {
		return fmt.Errorf("Template for frontpage and list is not exist")
	}

	// Prepare layout
	baseLayout := Layout{
		WebsiteTitle:  rd.Config.Title,
		WebsiteOwner:  rd.Config.Owner,
		ContentTitle:  rd.Config.Title,
		ContentDesc:   rd.Config.Description,
		ContentAuthor: rd.Config.Owner,
		Pages:         rd.Pages,
	}

	frontPage := List{
		Layout:      baseLayout,
		Path:        "/posts",
		CurrentPage: pageNumber,
		MaxPage:     rd.getMaxPagination(rd.Posts),
		Posts:       rd.getListPosts(rd.Posts, pageNumber),
		Categories:  rd.Categories,
		Tags:        rd.Tags,
	}

	// Execute templates
	tpl, err := template.New("").Funcs(funcsMap).ParseFiles(templates...)
	if err != nil {
		return err
	}

	return tpl.ExecuteTemplate(dst, activeTemplate, &frontPage)
}

// RenderList renders list template.
func (rd Renderer) RenderList(listType ListType, groupName string, pageNumber int, dst io.Writer) error {
	// Make sure config file is valid
	err := rd.validateConfig()
	if err != nil {
		return err
	}

	// Prepare templates
	themeDir := rd.Config.Theme
	tplList := fp.Join(themeDir, "list.html")
	if !fileExists(tplList) {
		return fmt.Errorf("Template for list is not exist")
	}

	templates, err := rd.getBaseTemplates()
	if err != nil {
		return err
	}

	templates = append(templates, tplList)

	// Filter posts by group
	filterCategory := func(post model.Post) bool {
		return post.Category == groupName
	}

	filterTag := func(post model.Post) bool {
		for _, tag := range post.Tags {
			if tag == groupName {
				return true
			}
		}
		return false
	}

	posts := []model.Post{}
	if listType == DEFAULT {
		posts = rd.Posts
	} else {
		filter := filterTag
		if listType == CATEGORY {
			filter = filterCategory
			if groupName == "uncategorized" {
				groupName = ""
			}
		}

		for _, post := range rd.Posts {
			if filter(post) {
				posts = append(posts, post)
			}
		}
	}

	// Prepare layout
	baseLayout := Layout{
		WebsiteTitle: rd.Config.Title,
		WebsiteOwner: rd.Config.Owner,
		ContentTitle: groupName,
		ContentDesc:  rd.Config.Description,
		Pages:        rd.Pages,
	}

	listPath := "/posts"
	if listType == CATEGORY {
		listPath = path.Join("/category", groupName)
	} else if listType == TAG {
		listPath = path.Join("/tag", groupName)
	}

	list := List{
		Layout:      baseLayout,
		Type:        listType,
		Path:        listPath,
		CurrentPage: pageNumber,
		MaxPage:     rd.getMaxPagination(posts),
		Posts:       rd.getListPosts(posts, pageNumber),
	}

	// Execute templates
	tpl, err := template.New("").Funcs(funcsMap).ParseFiles(templates...)
	if err != nil {
		return err
	}

	return tpl.ExecuteTemplate(dst, "list.html", &list)
}

// RenderPage renders page template.
func (rd Renderer) RenderPage(page model.Page, dst io.Writer) error {
	// Make sure config file is valid
	err := rd.validateConfig()
	if err != nil {
		return err
	}

	// Prepare templates
	themeDir := rd.Config.Theme
	tplPage := fp.Join(themeDir, "page.html")
	if !fileExists(tplPage) {
		return fmt.Errorf("Template for page is not exist")
	}

	templates, err := rd.getBaseTemplates()
	if err != nil {
		return err
	}

	templates = append(templates, tplPage)

	// Open index file
	content, err := readIndexFile(page.Path)
	if err != nil {
		return err
	}

	content = removeMetadata(content)
	htmlContent := blackfriday.Run(content)

	// Prepare layout
	baseLayout := Layout{
		WebsiteTitle: rd.Config.Title,
		WebsiteOwner: rd.Config.Owner,
		ContentTitle: page.Title,
		ContentDesc:  page.Excerpt,
		Pages:        rd.Pages,
	}

	pageLayout := Page{
		Layout:    baseLayout,
		Thumbnail: page.Thumbnail,
		HTML:      template.HTML(htmlContent),
	}

	// Execute templates
	tpl, err := template.New("").Funcs(funcsMap).ParseFiles(templates...)
	if err != nil {
		return err
	}

	return tpl.ExecuteTemplate(dst, "page.html", &pageLayout)
}

// RenderPost renders post template.
func (rd Renderer) RenderPost(post, olderPost, newerPost model.Post, dst io.Writer) error {
	// Make sure config file is valid
	err := rd.validateConfig()
	if err != nil {
		return err
	}

	// Prepare templates
	themeDir := rd.Config.Theme
	tplPost := fp.Join(themeDir, "post.html")
	if !fileExists(tplPost) {
		return fmt.Errorf("Template for post is not exist")
	}

	templates, err := rd.getBaseTemplates()
	if err != nil {
		return err
	}

	templates = append(templates, tplPost)

	// Convert category and tags of post into Group
	category := model.Group{
		Name: post.Category,
		Path: fp.Join("/", "category", post.Category),
	}

	if category.Name == "" {
		category.Path = fp.Join("/", "category", "uncategorized")
	}

	tags := []model.Group{}
	for _, tag := range post.Tags {
		tags = append(tags, model.Group{
			Name: tag,
			Path: fp.Join("/", "tag", tag),
		})
	}

	// Check author
	if post.Author == "" {
		post.Author = rd.Config.Owner
	}

	// Open index file
	content, err := readIndexFile(post.Path)
	if err != nil {
		return err
	}

	content = removeMetadata(content)
	html := blackfriday.Run(content)

	// Prepare layout
	baseLayout := Layout{
		WebsiteTitle:  rd.Config.Title,
		WebsiteOwner:  rd.Config.Owner,
		ContentTitle:  post.Title,
		ContentDesc:   post.Excerpt,
		ContentAuthor: post.Author,
		Pages:         rd.Pages,
	}

	postLayout := Post{
		Layout:    baseLayout,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,
		Category:  category,
		Tags:      tags,
		Thumbnail: post.Thumbnail,
		HTML:      template.HTML(html),
		Older:     olderPost,
		Newer:     newerPost,
	}

	// Execute templates
	tpl, err := template.New("").Funcs(funcsMap).ParseFiles(templates...)
	if err != nil {
		return err
	}

	return tpl.ExecuteTemplate(dst, "post.html", &postLayout)
}

func (rd Renderer) validateConfig() error {
	if rd.Config.BaseURL == "" {
		return fmt.Errorf("No base URL set in configuration file")
	}

	if rd.Config.Theme == "" {
		return fmt.Errorf("No theme specified in configuration file")
	}

	return nil
}

func (rd Renderer) getBaseTemplates() ([]string, error) {
	items, err := ioutil.ReadDir(rd.Config.Theme)
	if err != nil {
		return []string{}, err
	}

	templates := []string{}
	for _, item := range items {
		if item.IsDir() {
			continue
		}

		if strings.HasSuffix(item.Name(), ".html") && strings.HasPrefix(item.Name(), "_") {
			templates = append(templates, fp.Join(rd.Config.Theme, item.Name()))
		}
	}

	return templates, nil
}

func (rd Renderer) getMaxPagination(posts []model.Post) int {
	nPosts := len(posts)
	pageLength := rd.Config.Pagination

	fMaxPage := math.Ceil(float64(nPosts) / float64(pageLength))
	return int(fMaxPage)
}

func (rd Renderer) getListPosts(posts []model.Post, pageNumber int) []model.Post {
	nPosts := len(posts)
	pageLength := rd.Config.Pagination

	start := (pageNumber - 1) * pageLength
	end := start + pageLength
	if end > nPosts {
		end = nPosts
	}

	return posts[start:end]
}
