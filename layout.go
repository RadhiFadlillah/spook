package main

import "math"

// NewLayoutIndex returns new layout for index page
func NewLayoutIndex(page int, config Config, parsed ParsedAllPosts) LayoutIndex {
	// Create base layout
	baseLayout := Layout{
		WebsiteTitle:  config.Title,
		WebsiteOwner:  config.Owner,
		ContentTitle:  config.Title,
		ContentDesc:   config.Description,
		ContentAuthor: config.Owner,
		Categories:    parsed.Categories,
		Tags:          parsed.Tags,
	}

	// Calculate pagination
	nPosts := len(parsed.Posts)
	fMaxPage := math.Floor(float64(nPosts)/float64(config.Pagination) + 0.5)

	// Only pick posts for this page
	start := (page - 1) * config.Pagination
	end := start + config.Pagination
	if end > nPosts {
		end = nPosts
	}
	posts := parsed.Posts[start:end]

	return LayoutIndex{
		Layout:      baseLayout,
		Posts:       posts,
		CurrentPage: page,
		MaxPage:     int(fMaxPage),
	}
}
