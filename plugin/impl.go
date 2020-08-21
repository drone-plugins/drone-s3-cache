// Copyright (c) 2020, the Drone Plugins project authors.
// Please see the AUTHORS file for details. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file.

package plugin

import (
	"fmt"
	"net/url"
	"os"
	pathutil "path"
	"strings"
	"time"

	"github.com/drone-plugins/drone-s3-cache/storage/s3"
	"github.com/drone/drone-cache-lib/archive/util"
	"github.com/drone/drone-cache-lib/cache"
	"github.com/drone/drone-cache-lib/storage"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Settings for the plugin.
type Settings struct {
	Mode         string
	Root         string
	Filename     string
	Path         string
	FallbackPath string
	FlushPath    string
	FlushAge     int
	Mount        cli.StringSlice
	Restore      bool // DEPRECATED
	Rebuild      bool // DEPRECATED
	Flush        bool // DEPRECATED

	S3Options s3.Options
	mount     []string
}

const (
	restoreMode = "restore"
	rebuildMode = "rebuild"
	flushMode   = "flush"

	awsDomain   = "amazonaws.com"
	awsEndpoint = "https://s3." + awsDomain
)

// Validate handles the settings validation of the plugin.
func (p *Plugin) Validate() error {
	if err := p.validateMode(); err != nil {
		return err
	}
	return p.validateS3()
}

func (p *Plugin) validateMode() error {
	// Validate the mode
	mode := p.settings.Mode
	hasMode := p.settings.Rebuild || p.settings.Restore || p.settings.Flush
	if mode == "" {
		logrus.WithFields(logrus.Fields{
			"rebuild": p.settings.Rebuild,
			"restore": p.settings.Restore,
			"flush":   p.settings.Flush,
		}).Info("mode specified using boolean config")

		if !hasMode {
			return fmt.Errorf("no mode specified")
		}
		if multipleModesSpecified(p.settings.Rebuild, p.settings.Restore, p.settings.Flush) {
			return fmt.Errorf("multiple modes specified")
		}

		if p.settings.Rebuild {
			mode = rebuildMode
		} else if p.settings.Restore {
			mode = restoreMode
		} else {
			mode = flushMode
		}
	} else {
		if hasMode {
			return fmt.Errorf("mode specified multiple ways")
		}

		if mode != rebuildMode && mode != restoreMode && mode != flushMode {
			return fmt.Errorf("invalid mode %s specified", mode)
		}
	}

	logrus.WithField("mode", mode).Info("using mode")
	p.settings.Mode = mode

	if p.settings.Filename == "" {
		logrus.Debug("using default filename")
		p.settings.Filename = "archive.tar"
	}
	logrus.WithField("filename", p.settings.Filename).Debug("using filename")

	// Validate mode settings
	if mode != flushMode {
		if p.settings.Path == "" {
			logrus.WithFields(logrus.Fields{
				"repo.owner":    p.pipeline.Repo.Owner,
				"repo.name":     p.pipeline.Repo.Name,
				"commit.branch": p.pipeline.Commit.Branch,
			}).Debug("creating default path")
			p.settings.Path = fmt.Sprintf(
				"%s/%s/%s",
				p.pipeline.Repo.Owner,
				p.pipeline.Repo.Name,
				p.pipeline.Commit.Branch,
			)
		}
		logrus.WithField("path", p.settings.Path).Debug("using path")

		if mode == rebuildMode {
			mount := p.settings.Mount.Value()
			if len(mount) == 0 {
				return fmt.Errorf("cache not specified")
			}
			p.settings.mount = mount
		} else {
			if p.settings.FallbackPath == "" {
				logrus.WithFields(logrus.Fields{
					"repo.owner":  p.pipeline.Repo.Owner,
					"repo.name":   p.pipeline.Repo.Name,
					"repo.branch": p.pipeline.Repo.Branch,
				}).Debug("creating default fallback path")
				p.settings.FallbackPath = fmt.Sprintf(
					"%s/%s/%s",
					p.pipeline.Repo.Owner,
					p.pipeline.Repo.Name,
					p.pipeline.Repo.Branch,
				)
			}
			logrus.WithField("path", p.settings.FallbackPath).Debug("using path as fallback")
		}
	} else {
		if p.settings.FlushPath == "" {
			logrus.WithFields(logrus.Fields{
				"repo.owner": p.pipeline.Repo.Owner,
				"repo.name":  p.pipeline.Repo.Name,
			}).Debug("creating default flush path")
			p.settings.FlushPath = fmt.Sprintf(
				"%s/%s",
				p.pipeline.Repo.Owner,
				p.pipeline.Repo.Name,
			)
		}
		logrus.WithField("path", p.settings.FlushPath).Debug("using path when flushing")
	}

	return nil
}

func (p *Plugin) validateS3() error {
	// Validate the endpoint
	endpoint := p.settings.S3Options.Endpoint
	isAWS := false
	bucket := ""
	region := ""

	if endpoint == "" {
		endpoint = awsEndpoint
	}

	s3url, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("could not parse endpoint %s", endpoint)
	}

	// Check if additional information is encoded in the endpoint
	if h := s3url.Hostname(); strings.HasSuffix(h, awsDomain) {
		isAWS = true

		// Sub-domains can specify region and bucket
		d := strings.Split(h, ".")
		s3sub := 0

		switch len(d) {
		case 5:
			// Virtual hosted style access
			// https://bucket-name.s3.Region.amazonaws.com/
			logrus.WithField("host", h).Debug("using virtual host style access")
			bucket = d[0]
			s3sub = 1
			region = d[2]
		case 4:
			// Path-style access
			// https://s3.Region.amazonaws.com/bucket-name
			logrus.WithField("host", h).Debug("using path style access")
			bucket = s3url.Path
			s3sub = 0
			region = d[1]
		case 3:
			// Just default url https://s3.Region.amazonaws.com
		default:
			return fmt.Errorf("unknown aws domain for url %s", endpoint)
		}

		if d[s3sub] != "s3" {
			return fmt.Errorf("unknown aws domain for url %s", endpoint)
		}

		// Normalize endpoint
		endpoint = awsEndpoint
		s3url, _ = url.Parse(endpoint)
	}

	// Check for s3 scheme
	if s3url.Scheme == "s3" {
		logrus.WithField("endpoint", endpoint).Debug("using s3 url")
		bucket = s3url.Hostname()

		// Normalize endpoint
		endpoint = awsEndpoint
		s3url, _ = url.Parse(endpoint)
	}

	var useSSL bool
	switch s3url.Scheme {
	case "https":
		endpoint = endpoint[8:]
		useSSL = true
	case "http":
		endpoint = endpoint[7:]
		useSSL = false
	default:
		return fmt.Errorf("unknown scheme for endpoint %s", endpoint)
	}

	if bucket != "" {
		logrus.WithField("bucket", bucket).Info("bucket found in S3 endpoint")
		if p.settings.Root != "" {
			return fmt.Errorf("bucket %s already specified in endpoint remove from root", bucket)
		}
		p.settings.Root = bucket
	}

	if region != "" {
		logrus.WithField("region", region).Info("region found in S3 endpoint")
		if p.settings.S3Options.Region != "" {
			return fmt.Errorf("region %s already specified in endpoint remove from config", region)
		}
		p.settings.S3Options.Region = region
	}
	s3Opts := p.settings.S3Options

	if (s3Opts.Access != "" || s3Opts.Secret != "") && s3Opts.FileCredentialsPath != "" {
		return fmt.Errorf("only one credentials method should be used. Use either access-key and secret-key OR the credentials file")
	}

	if s3Opts.FileCredentialsPath != "" {
		if _, err := os.Stat(s3Opts.FileCredentialsPath); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", s3Opts.FileCredentialsPath)
		}
	}

	logrus.WithFields(logrus.Fields{
		"endpoint": endpoint,
		"use-ssl":  useSSL,
	}).Info("using S3 endpoint")
	p.settings.S3Options.Endpoint = endpoint
	p.settings.S3Options.UseSSL = useSSL

	if isAWS && p.settings.Root == "" {
		return fmt.Errorf("no aws bucket specified in root or endpoint")
	}

	return nil
}

// Execute provides the implementation of the plugin.
func (p *Plugin) Execute() error {
	at, err := util.FromFilename(p.settings.Filename)
	if err != nil {
		return err
	}

	st, err := s3.New(&p.settings.S3Options)
	if err != nil {
		return err
	}

	c := cache.New(st, at)

	if p.settings.Mode == rebuildMode {
		path := cleanPath(p.settings.Root, p.settings.Path, p.settings.Filename)
		logrus.WithFields(logrus.Fields{
			"path": path,
		}).Info("rebuilding cache")
		err = c.Rebuild(p.settings.mount, path)

		if err == nil {
			logrus.Infof("cache rebuilt")
		}
	} else if p.settings.Mode == restoreMode {
		path := cleanPath(p.settings.Root, p.settings.Path, p.settings.Filename)
		fallbackPath := cleanPath(p.settings.Root, p.settings.FallbackPath, p.settings.Filename)

		logrus.WithFields(logrus.Fields{
			"path":     path,
			"fallback": fallbackPath,
		}).Info("restoring cache")
		err = c.Restore(path, fallbackPath)

		if err == nil {
			logrus.Info("cache restored")
		}
	} else /* p.settings.Mode == flushMode */ {
		flushPath := cleanPath(p.settings.Root, p.settings.FlushPath)

		logrus.WithFields(logrus.Fields{
			"path":    flushPath,
			"max-age": p.settings.FlushAge,
		}).Info("flushing cache")
		f := cache.NewFlusher(st, genIsExpired(p.settings.FlushAge))
		err = f.Flush(flushPath)

		if err == nil {
			logrus.Info("Cache flushed")
		}
	}

	return err
}

func genIsExpired(age int) cache.DirtyFunc {
	return func(file storage.FileEntry) bool {
		// Check if older than "age" days
		return file.LastModified.Before(time.Now().AddDate(0, 0, age*-1))
	}
}

func cleanPath(paths ...string) string {
	return pathutil.Clean(pathutil.Join(paths...))
}

func multipleModesSpecified(bools ...bool) bool {
	var b bool
	for _, v := range bools {
		if b && b == v {
			return true
		}

		if v {
			b = true
		}
	}

	return false
}
