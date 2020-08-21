// Copyright (c) 2020, the Drone Plugins project authors.
// Please see the AUTHORS file for details. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be
// found in the LICENSE file.

package main

import (
	"github.com/drone-plugins/drone-s3-cache/plugin"
	"github.com/urfave/cli/v2"
)

// settingsFlags has the cli.Flags for the plugin.Settings.
func settingsFlags(settings *plugin.Settings) []cli.Flag {
	return []cli.Flag{
		// Cache information

		&cli.StringFlag{
			Name:        "mode",
			Usage:       "set plugin mode (rebuild,restore,flush)",
			EnvVars:     []string{"PLUGIN_MODE"},
			Destination: &settings.Mode,
		},
		&cli.StringFlag{
			Name:        "filename",
			Usage:       "filename for the cache archive",
			EnvVars:     []string{"PLUGIN_FILENAME"},
			Destination: &settings.Filename,
		},
		&cli.StringFlag{
			Name:        "root",
			Usage:       "storage root of cache files",
			EnvVars:     []string{"PLUGIN_ROOT"},
			Destination: &settings.Root,
		},
		&cli.StringFlag{
			Name:        "path",
			Usage:       "path to cache files relative to root",
			EnvVars:     []string{"PLUGIN_PATH"},
			Destination: &settings.Path,
		},
		&cli.StringFlag{
			Name:        "fallback-path",
			Usage:       "path to default cache files relative to the root",
			EnvVars:     []string{"PLUGIN_FALLBACK_PATH"},
			Destination: &settings.FallbackPath,
		},
		&cli.StringFlag{
			Name:        "flush-path",
			Usage:       "path to flushable cache files relative to the root",
			EnvVars:     []string{"PLUGIN_FLUSH_PATH"},
			Destination: &settings.FlushPath,
		},
		&cli.StringSliceFlag{
			Name:        "mount",
			Usage:       "directories to cache",
			EnvVars:     []string{"PLUGIN_MOUNT"},
			Destination: &settings.Mount,
		},
		&cli.IntFlag{
			Name:        "flush-age",
			Usage:       "flush cache files older than # days",
			EnvVars:     []string{"PLUGIN_FLUSH_AGE"},
			Value:       30,
			Destination: &settings.FlushAge,
		},

		// Cache information (deprecated)

		&cli.BoolFlag{
			Name:        "rebuild",
			Usage:       "rebuild the cache directories",
			EnvVars:     []string{"PLUGIN_REBUILD"},
			Destination: &settings.Rebuild,
		},
		&cli.BoolFlag{
			Name:        "restore",
			Usage:       "restore the cache directories",
			EnvVars:     []string{"PLUGIN_RESTORE"},
			Destination: &settings.Restore,
		},
		&cli.BoolFlag{
			Name:        "flush",
			Usage:       "flush the cache",
			EnvVars:     []string{"PLUGIN_FLUSH"},
			Destination: &settings.Flush,
		},

		// S3 information

		&cli.StringFlag{
			Name:        "endpoint",
			Usage:       "s3 endpoint",
			EnvVars:     []string{"PLUGIN_SERVER", "PLUGIN_ENDPOINT", "CACHE_S3_ENDPOINT", "CACHE_S3_SERVER", "S3_ENDPOINT"},
			Destination: &settings.S3Options.Endpoint,
		},
		&cli.StringFlag{
			Name:        "accelerated-endpoint",
			Usage:       "s3 accelerated endpoint",
			EnvVars:     []string{"PLUGIN_ACCELERATED_ENDPOINT", "CACHE_S3_ACCELERATED_ENDPOINT"},
			Destination: &settings.S3Options.AcceleratedEndpoint,
		},
		&cli.StringFlag{
			Name:        "access-key",
			Usage:       "s3 access key",
			EnvVars:     []string{"PLUGIN_ACCESS_KEY", "CACHE_S3_ACCESS_KEY", "AWS_ACCESS_KEY_ID"},
			Destination: &settings.S3Options.Access,
		},
		&cli.StringFlag{
			Name:        "secret-key",
			Usage:       "s3 secret key",
			EnvVars:     []string{"PLUGIN_SECRET_KEY", "CACHE_S3_SECRET_KEY", "AWS_SECRET_ACCESS_KEY"},
			Destination: &settings.S3Options.Secret,
		},
		&cli.StringFlag{
			Name:        "session-token",
			Usage:       "s3 session token",
			EnvVars:     []string{"PLUGIN_SESSION_TOKEN", "CACHE_S3_SESSION_TOKEN", "AWS_SESSION_TOKEN"},
			Destination: &settings.S3Options.Token,
		},
		&cli.StringFlag{
			Name:        "region",
			Usage:       "s3 region",
			EnvVars:     []string{"PLUGIN_REGION", "CACHE_S3_REGION"},
			Destination: &settings.S3Options.Region,
		},
		&cli.StringFlag{
			Name:        "file-credentials",
			Usage:       "path to s3 credentials file",
			EnvVars:     []string{"PLUGIN_FILE_CREDENTIALS_PATH", "CACHE_FILE_CREDENTIALS_PATH", "AWS_SHARED_CREDENTIALS_FILE"},
			Destination: &settings.S3Options.FileCredentials,
		},
		&cli.StringFlag{
			Name:        "profile",
			Usage:       "s3 profile name",
			EnvVars:     []string{"PLUGIN_PROFILE", "CACHE_S3_PROFILE", "AWS_PROFILE"},
			Destination: &settings.S3Options.Profile,
		},
	}
}
