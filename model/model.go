package model

// Config is data of main configuration file
type Config struct {
	Title       string
	Description string
	Owner       string
	Pagination  int
	Theme       string
	PublishDir  string
}

// Group is keyword for grouping several posts
type Group struct {
	Name   string
	Path   string
	NPosts int
}

// Page is a static standalone content
type Page struct {
	Title     string
	Excerpt   string
	Path      string `toml:"-"`
	Thumbnail string `toml:"-"`
}

// Post is the content that listed in chronological order
type Post struct {
	Title     string
	Excerpt   string
	CreatedAt string
	UpdatedAt string
	Category  string
	Tags      []string
	Author    string
	Path      string `toml:"-"`
	Thumbnail string `toml:"-"`
}
