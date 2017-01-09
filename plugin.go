package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/drone/drone-cache-lib/cache"
	"github.com/drone/drone-cache-lib/storage"
)

type Plugin struct {
	Filename     string
	Path         string
	FallbackPath string
	FlushPath    string
	Mode         string
	Flush        bool
	FlushAge     int
	Mount        []string

	Storage storage.Storage
}

const (
	RestoreMode = "restore"
	RebuildMode = "rebuild"
)

// Exec runs the plugin
func (p *Plugin) Exec() error {
	var err error

	c := cache.NewDefault(p.Storage)

	path := p.Path + p.Filename
	fallbackPath := p.FallbackPath + p.Filename

	if p.Mode == RebuildMode {
		log.Infof("Rebuilding cache at %s", path)
		err = c.Rebuild(p.Mount, path)

		if err == nil {
			log.Infof("Cache rebuilt")
		} else {
			return err
		}
	} else {
		log.Infof("Restoring cache at %s", path)
		err = c.Restore(path, fallbackPath)

		if err == nil {
			log.Info("Cache restored")
		} else {
			return err
		}
	}

	if p.Flush {
		f := cache.NewFlusher(p.Storage, genIsExpired(p.FlushAge))
		err = f.Flush(p.FlushPath)
	}

	return err
}

func genIsExpired(age int) cache.DirtyFunc {
	return func(file storage.FileEntry) bool {
		// Check if older then 30 days
		if file.LastModified.Before(time.Now().AddDate(0, 0, age * -1)) {
			return true
		}

		// No match
		return false
	}
}
