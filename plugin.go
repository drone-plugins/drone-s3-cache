package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/drone/drone-cache-lib/cache"
	"github.com/drone/drone-cache-lib/storage"
	"github.com/drone/drone-cache-lib/archive"
	"github.com/drone/drone-cache-lib/archive/tar"
	"github.com/drone/drone-cache-lib/archive/tgz"
)

type Plugin struct {
	Filename     string
	Path         string
	FallbackPath string
	ArchiveType  string
	FlushPath    string
	Mode         string
	FlushAge     int
	Mount        []string

	Storage storage.Storage
}

const (
	RestoreMode = "restore"
	RebuildMode = "rebuild"
	FlushMode = "flush"
)

// Exec runs the plugin
func (p *Plugin) Exec() error {
	var err error
	var ta archive.Archive

	switch at := p.ArchiveType; at {
	case "tgz":
		ta = tgz.New()
	default:
		ta = tar.New()
	}

	c := cache.New(p.Storage, ta)

	path := p.Path + p.Filename
	fallbackPath := p.FallbackPath + p.Filename

	if p.Mode == RebuildMode {
		log.Infof("Rebuilding cache at %s", path)
		err = c.Rebuild(p.Mount, path)

		if err == nil {
			log.Infof("Cache rebuilt")
		}
	}

	if p.Mode == RestoreMode {
		log.Infof("Restoring cache at %s", path)
		err = c.Restore(path, fallbackPath)

		if err == nil {
			log.Info("Cache restored")
		}
	}

	if p.Mode == FlushMode {
		log.Infof("Flushing cache items older then %s days at %s", p.FlushAge, path)
		f := cache.NewFlusher(p.Storage, genIsExpired(p.FlushAge))
		err = f.Flush(p.FlushPath)

		if err == nil {
			log.Info("Cache flushed")
		}
	}

	return err
}

func genIsExpired(age int) cache.DirtyFunc {
	return func(file storage.FileEntry) bool {
		// Check if older then "age" days
		return file.LastModified.Before(time.Now().AddDate(0, 0, age * -1))
	}
}
