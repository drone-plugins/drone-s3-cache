package storage

import (
	"io"
	"time"
)

type FileEntry struct {
	Path         string
	Size         int64
	LastModified time.Time
}

// Storage is a place that files can be written to and read from.
type Storage interface {
	Get(p string, dst io.Writer) error
	Put(p string, src io.Reader) error
	List(p string) ([]FileEntry, error)
	Delete(p string) error
}
