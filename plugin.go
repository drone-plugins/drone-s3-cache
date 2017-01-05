package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/drone/drone-cache-lib/cache"
	"github.com/drone/drone-cache-lib/storage"
)

type Plugin struct {
	Filename     string
	Path         string
	FallbackPath string
	Mode         string
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
		}

		return err
	}

	log.Infof("Restoring cache at %s", path)
	err = c.Restore(path, fallbackPath)

	if err == nil {
		log.Info("Cache restored")
	}

	return err
}
