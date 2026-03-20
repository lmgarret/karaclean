package engine

import "time"

// Bookmark is the domain model for rule evaluation.
// Mapped from Karakeep API response by the karakeep client wrapper.
type Bookmark struct {
	ID         string
	CreatedAt  time.Time
	Archived   bool
	Favourited bool
	Source     string   // one of: rss, web, api, mobile, extension, cli, import, singlefile
	Tags       []string // tag names extracted from tag objects
	Note       string   // user's personal note (empty string if none)
	Size       int64    // content size in bytes (0 means unknown/not applicable)
}
