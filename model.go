package main

import "time"

// Config is data of main configuration file
type Config struct {
	Title       string
	Description string
	BaseURL     string
	Pagination  int
}

// Group is keyword for grouping several posts
type Group struct {
	URL    string
	Name   string
	NPosts int
}

// Page is a static standalone content
type Page struct {
	URL     string
	Title   string
	Content string
}

// Post is the content that listed in chronological order
type Post struct {
	URL       string
	Title     string
	Excerpt   string
	CreatedAt string
	UpdatedAt string
	Category  string
	Tags      []string
	Content   string
}

// Pagination is data of the pagination
type Pagination struct {
	Length  int
	Current int
	Max     int
}

// Layout is base layout of the website
type Layout struct {
	WebsiteTitle string
	ContentTitle string
	ContentDesc  string
	ListPage     []Page
	ListCategory []Group
	ListTag      []Group
}

// IndexLayout is layout that used in index and list template
type IndexLayout struct {
	Layout
	ListPost   []Post
	Pagination Pagination
}

// PageLayout is layout that used in page
type PageLayout struct {
	Layout
	Content string
}

// PostLayout is layout that used in post
type PostLayout struct {
	Layout
	CreatedAt time.Time
	UpdatedAt time.Time
	Category  Group
	Tags      []Group
	Content   string
	Previous  Post
	Next      Post
}
