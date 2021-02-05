package main

import (
	"errors"
	"fmt"
	"os"
	pathutil "path"
	"strconv"
	"strings"

	"github.com/drone-plugins/drone-s3-cache/storage/s3"
	"github.com/drone/drone-cache-lib/storage"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version = "unknown"
)

func main() {
	app := cli.NewApp()
	app.Name = "cache plugin"
	app.Usage = "cache plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{
		// Cache information

		cli.StringFlag{
			Name:   "filename",
			Usage:  "filename for the cache",
			EnvVar: "PLUGIN_FILENAME",
		},
		cli.StringFlag{
			Name:   "root",
			Usage:  "root",
			EnvVar: "PLUGIN_ROOT",
		},
		cli.StringFlag{
			Name:   "path",
			Usage:  "path",
			EnvVar: "PLUGIN_PATH",
		},
		cli.StringFlag{
			Name:   "fallback_path",
			Usage:  "fallback_path",
			EnvVar: "PLUGIN_FALLBACK_PATH",
		},
		cli.StringSliceFlag{
			Name:   "mount",
			Usage:  "cache directories",
			EnvVar: "PLUGIN_MOUNT",
		},
		cli.BoolFlag{
			Name:   "rebuild",
			Usage:  "rebuild the cache directories",
			EnvVar: "PLUGIN_REBUILD",
		},
		cli.BoolFlag{
			Name:   "restore",
			Usage:  "restore the cache directories",
			EnvVar: "PLUGIN_RESTORE",
		},
		cli.BoolFlag{
			Name:   "flush",
			Usage:  "flush the cache",
			EnvVar: "PLUGIN_FLUSH",
		},
		cli.StringFlag{
			Name:   "flush_age",
			Usage:  "flush cache files older than # days",
			EnvVar: "PLUGIN_FLUSH_AGE",
			Value:  "30",
		},
		cli.StringFlag{
			Name:   "flush_path",
			Usage:  "path to search for flushable cache files",
			EnvVar: "PLUGIN_FLUSH_PATH",
		},
		cli.StringFlag{
			Name:   "workdir",
			Usage:  "path where the cache will be extracted to",
			EnvVar: "PLUGIN_WORKDIR",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug plugin output",
			EnvVar: "PLUGIN_DEBUG",
		},

		// Build information (for setting defaults)

		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "repo.branch",
			Value:  "master",
			Usage:  "repository default branch",
			EnvVar: "DRONE_REPO_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},

		// S3 information

		cli.StringFlag{
			Name:   "server",
			Usage:  "s3 server",
			EnvVar: "PLUGIN_SERVER,PLUGIN_ENDPOINT,CACHE_S3_ENDPOINT,CACHE_S3_SERVER,S3_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "accelerated-endpoint",
			Usage:  "s3 accelerated endpoint",
			EnvVar: "PLUGIN_ACCELERATED_ENDPOINT,CACHE_S3_ACCELERATED_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "access-key",
			Usage:  "s3 access key",
			EnvVar: "PLUGIN_ACCESS_KEY,CACHE_S3_ACCESS_KEY,AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "s3 secret key",
			EnvVar: "PLUGIN_SECRET_KEY,CACHE_S3_SECRET_KEY,AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "s3 region",
			EnvVar: "PLUGIN_REGION,CACHE_S3_REGION",
		},
		cli.StringFlag{
			Name:   "ca_cert",
			Usage:  "ca cert to connect to s3 server",
			EnvVar: "PLUGIN_CA_CERT,CACHE_S3_CA_CERT",
		},
		cli.StringFlag{
			Name:   "ca_cert_path",
			Usage:  "ca cert to connect to s3 server",
			EnvVar: "PLUGIN_CA_CERT_PATH,CACHE_S3_CA_CERT_PATH",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	// Set the logging level
	if c.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	// Determine the mode for the plugin
	rebuild := c.Bool("rebuild")
	restore := c.Bool("restore")
	flush := c.Bool("flush")

	if isMultipleModes(rebuild, restore, flush) {
		return errors.New("Must use a single mode: rebuild, restore or flush")
	} else if !rebuild && !restore && !flush {
		return errors.New("No action specified")
	}

	var mode string
	var mount []string

	if rebuild {
		// Look for the mount points to rebuild
		mount = c.StringSlice("mount")

		if len(mount) == 0 {
			return errors.New("No mounts specified")
		}

		mode = RebuildMode
	} else if flush {
		mode = FlushMode
	} else {
		mode = RestoreMode
	}

	// Get the working directory
	workdir := c.String("workdir")

	// Get the root path prefix to place the cache files
	root := c.GlobalString("root")

	// Get the path to place the cache files
	path := c.GlobalString("path")

	// Defaults to <owner>/<repo>/<branch>/
	if len(path) == 0 {
		log.Info("No path specified. Creating default")

		path = fmt.Sprintf(
			"%s/%s/%s",
			c.String("repo.owner"),
			c.String("repo.name"),
			c.String("commit.branch"),
		)

		path = prefixRoot(root, path)
	}

	// Get the fallback path to retrieve the cache files
	fallbackPath := c.GlobalString("fallback_path")

	// Defaults to <owner>/<repo>/master/
	if len(fallbackPath) == 0 {
		log.Info("No fallback_path specified. Creating default")

		fallbackPath = fmt.Sprintf(
			"%s/%s/%s",
			c.String("repo.owner"),
			c.String("repo.name"),
			c.String("repo.branch"),
		)

		fallbackPath = prefixRoot(root, fallbackPath)
	}

	// Get the flush path to flush the cache files from
	flushPath := c.GlobalString("flush_path")

	// Defaults to <owner>/<repo>/
	if len(flushPath) == 0 {
		log.Info("No flush_path specified. Creating default")

		flushPath = fmt.Sprintf(
			"%s/%s",
			c.String("repo.owner"),
			c.String("repo.name"),
		)

		flushPath = prefixRoot(root, flushPath)
	}

	// Get the filename
	filename := c.GlobalString("filename")

	if len(filename) == 0 {
		log.Info("No filename specified. Creating default")

		filename = "archive.tar"
	}

	s, err := s3Storage(c)

	if err != nil {
		return err
	}

	flushAge, err := strconv.Atoi(c.String("flush_age"))


	if err != nil {
		return err
	}

	p := &Plugin{
		Filename:     filename,
		Path:         path,
		FallbackPath: fallbackPath,
		FlushPath:    flushPath,
		Mode:         mode,
		FlushAge:     flushAge,
		Mount:        mount,
		Storage:      s,
		Cacert:       c.String("ca_cert"),
		CacertPath:   c.String("ca_cert_path"),
		Workdir:      workdir,
	}

	return p.Exec()
}

func s3Storage(c *cli.Context) (storage.Storage, error) {
	// Get the endpoint
	server := c.String("server")

	var endpoint string
	var useSSL bool

	if len(server) > 0 {
		useSSL = strings.HasPrefix(server, "https://")

		if !useSSL {
			if !strings.HasPrefix(server, "http://") {
				return nil, fmt.Errorf("Invalid server %s. Needs to be a HTTP URI", server)
			}

			endpoint = server[7:]
		} else {
			endpoint = server[8:]
		}
	} else {
		endpoint = "s3.amazonaws.com"
		useSSL = true
	}

	return s3.New(&s3.Options{
		Endpoint:            endpoint,
		AcceleratedEndpoint: c.String("accelerated-endpoint"),
		Access:              c.String("access-key"),
		Secret:              c.String("secret-key"),
		Region:              c.String("region"),
		UseSSL:              useSSL,
	})
}

func isMultipleModes(bools ...bool) bool {
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

func prefixRoot(root, path string) string {
	return pathutil.Clean(fmt.Sprintf("/%s/%s", root, path))
}
