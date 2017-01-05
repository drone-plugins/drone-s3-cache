package s3

import (
	"fmt"
	"io"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/drone/drone-cache-lib/storage"
	"github.com/dustin/go-humanize"
	"github.com/minio/minio-go"
)

// Options contains configuration for the S3 connection.
type Options struct {
	Endpoint   string
	Key        string
	Secret     string
	Encryption string
	Access     string

	// us-east-1
	// us-west-1
	// us-west-2
	// eu-west-1
	// ap-southeast-1
	// ap-southeast-2
	// ap-northeast-1
	// sa-east-1
	Region string

	// Use path style instead of domain style.
	//
	// Should be true for minio and false for AWS.
	PathStyle bool

	UseSSL bool
}

type s3Storage struct {
	client *minio.Client
	opts   *Options
}

// NewS3Storage creates an implementation of Storage with S3 as the backend.
func New(opts *Options) (storage.Storage, error) {
	client, err := minio.New(opts.Endpoint, opts.Access, opts.Secret, opts.UseSSL)

	if err != nil {
		return nil, err
	}

	return &s3Storage{
		client: client,
		opts:   opts,
	}, nil
}

func (s *s3Storage) Get(p string, dst io.Writer) error {
	bucket, key := splitBucket(p)

	if len(bucket) == 0 || len(key) == 0 {
		return fmt.Errorf("Invalid path %s", p)
	}

	log.Infof("Retrieving file in %s at %s", bucket, key)

	exists, err := s.client.BucketExists(bucket)

	if !exists {
		return err
	}

	object, err := s.client.GetObject(bucket, key)
	if err != nil {
		return err
	}

	log.Infof("Copying object from the server")

	numBytes, err := io.Copy(dst, object)

	if err != nil {
		return err
	}

	log.Infof("Downloaded %s from server", humanize.Bytes(uint64(numBytes)))

	return nil
}

func (s *s3Storage) Put(p string, src io.Reader) error {
	bucket, key := splitBucket(p)

	log.Infof("Uploading to bucket %s at %s", bucket, key)

	if len(bucket) == 0 || len(key) == 0 {
		return fmt.Errorf("Invalid path %s", p)
	}

	exists, err := s.client.BucketExists(bucket)

	if !exists || err != nil {
		if err = s.client.MakeBucket(bucket, s.opts.Region); err != nil {
			return err
		}
		log.Infof("Bucket %s created", bucket)
	} else {
		log.Infof("Bucket %s already exists", bucket)
	}

	log.Infof("Putting file in %s at %s", bucket, key)

	numBytes, err := s.client.PutObject(bucket, key, src, "application/tar")

	if err != nil {
		return err
	}

	log.Infof("Uploaded %s to server", humanize.Bytes(uint64(numBytes)))

	return nil
}

func (s *s3Storage) List(p string) ([]storage.FileEntry, error) {
	bucket, key := splitBucket(p)

	log.Infof("Retrieving object in bucket %s at %s", bucket, key)

	if len(bucket) == 0 || len(key) == 0 {
		return nil, fmt.Errorf("Invalid path %s", p)
	}

	exists, err := s.client.BucketExists(bucket)

	if err != nil {
		return nil, fmt.Errorf("%s does not exist: %s", p, err)
	}
	if !exists {
		return nil, fmt.Errorf("%s does not exist", p)
	}

	// Create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	var objects []storage.FileEntry
	isRecursive := true
	objectCh := s.client.ListObjectsV2(bucket, key, isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("Failed to retreive object %s: %s", object.Key, object.Err)
		}

		objects = append(objects, storage.FileEntry{
			Path: bucket + "/" + key + "/" + object.Key,
			Size: object.Size,
			LastModified: object.LastModified,
		})
	}

	return objects, nil
}

func (s *s3Storage) Delete(p string) error {
	return nil
}

func splitBucket(p string) (string, string) {
	// Remove initial forward slash
	full := strings.TrimPrefix(p, "/")

	// Get first index
	i := strings.Index(full, "/")

	if i != -1 && len(full) != i+1 {
		// Bucket names need to be all lower case for the key it doesnt matter
		return strings.ToLower(full[0:i]), full[i+1:]
	}

	return "", ""
}
