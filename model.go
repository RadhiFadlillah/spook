package main

import (
	"html/template"
	"time"
)

// Config is data of main configuration file
type Config struct {
	Title       string
	Description string
	Owner       string
	BaseURL     string
	Pagination  int
	Theme       string
}

// Group is keyword for grouping several posts
type Group struct {
	URL    string
	Name   string
	NPosts int
}

// Page is a static standalone content
type Page struct {
	Title string
	URL   string `toml:"-"`
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
	URL       string `toml:"-"`
	Content   string `toml:"-"`
}

//
// BENEATH THIS POINT IS MODEL FOR PARSER
// ++++++++++++++++++++++++++++++++++++++

// ParsedAllPosts is results of parsing all post
type ParsedAllPosts struct {
	Categories []Group
	Tags       []Group
	Posts      []Post
}

//
// BENEATH THIS POINT IS MODEL FOR LAYOUT BUILDER
// ++++++++++++++++++++++++++++++++++++++++++++++

// Layout is base layout of the website
type Layout struct {
	WebsiteTitle  string
	WebsiteOwner  string
	ContentTitle  string
	ContentDesc   string
	ContentAuthor string
	Categories    []Group
	Tags          []Group
	AllPosts      []Post
}

// LayoutIndex is layout that used in index template
type LayoutIndex struct {
	Layout
	Posts       []Post
	CurrentPage int
	MaxPage     int
}

// LayoutList is layout that used in list template
type LayoutList struct {
	Layout
	Type        string
	Posts       []Post
	CurrentPage int
	MaxPage     int
}

// LayoutPage is layout that used in single page
type LayoutPage struct {
	Layout
	HTML template.HTML
}

// LayoutPost is layout that used in post
type LayoutPost struct {
	Layout
	CreatedAt time.Time
	UpdatedAt time.Time
	Category  Group
	Tags      []Group
	HTML      template.HTML
	Previous  Post
	Next      Post
}
