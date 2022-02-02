package filesystems

import "time"

// FS ist the interface for file systems. In order to satisfy the interface, all functions must exist
type FS interface {
	Put(fileName, folder string) error
	Get(destination string, items ...string) error
	List(prefix string) ([]Listing, error)
	Delete(itemsToDelete []string) bool
}

// Listing describes a file on a remote file system
type Listing struct {
	Etag         string
	LastModified time.Time
	Key          string
	Size         float64
	IsDir        bool
}