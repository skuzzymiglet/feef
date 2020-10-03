package feeds

import (
	"crypto/sha256"
	"time"
)

type Author struct {
	Name  string
	Email string
}

type Feed struct {
	Title       string
	Link        string
	FeedLink    string
	Updated     time.Time
	Published   time.Time
	Description string
	Items       []Item
	Content     string
}

type Item struct {
	Parent   *Feed
	ID       [sha256.Size]byte // Use SHA256 sum, so all GUIDs are same format
	Title    string
	Link     string
	Pulished time.Time
	Content  string
}
