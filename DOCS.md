You can use the cache plugin to save and restore your build environment.  This
is useful if you are downloading and/or compiling packages that can be reused
across tests.

```yaml
pipeline:
  restore_from_cache:
    image: jmccann/drone-s3-cache:1
    url: http://minio.company.com
    access_key: myaccesskey
    secret_key: supersecretKey
    pull: true
    restore: true

  build:
    image: ruby:2.3
    commands:
      - bundle install --path .bundler
      - bundle exec rake

  rebuild_cache:
    image: jmccann/drone-s3-cache:1
    url: http://minio.company.com
    access_key: myaccesskey
    secret_key: supersecretKey
    rebuild: true
    mount:
      - .bundler

  flush_cache:
    image: jmccann/drone-s3-cache:1
    url: http://minio.company.com
    access_key: myaccesskey
    secret_key: supersecretKey
    flush: true
    flush_age: 14
```

Update the files/directories to what **YOU** want by modifying the `mount` key.

```diff
pipeline:
  restore_from_cache:
    image: jmccann/drone-s3-cache:1
    url: http://minio.company.com
    access_key: myaccesskey
    secret_key: supersecretKey
    pull: true
    restore: true

    build:
      image: ruby:2.3
      commands:
        - bundle install --path .bundler
        - bundle exec rake

  rebuild_cache:
    image: jmccann/drone-s3-cache:1
    url: http://minio.company.com
    access_key: myaccesskey
    secret_key: supersecretKey
    rebuild: true
    mount:
-     - .bundler
+     - <yourstuffhere>
+     - <morestuffhere>

  flush_cache:
    image: jmccann/drone-s3-cache:1
    url: http://minio.company.com
    access_key: myaccesskey
    secret_key: supersecretKey
    flush: true
    flush_age: 14
```

# Secrets

All plugins supports reading credentials from the Drone secret store. This is
strongly recommended instead of storing credentials in the pipeline configuration
in plain text. Please see the Drone [documentation]({{< secret-link >}}) to learn
more about secrets.

# Parameters

* `url`: The server url for your S3 instance
* `access_key`: The access key for your S3 instance
* `secret_key`: The secret key for your S3 instance
* `restore`: Restore the build environment from cache
* `rebuild`: Rebuild the cache from the build environemnt and specified `mount`s
* `flush`: Flush the cache of old cache items (please be sure to set this so we don't waste storage)
* `mount`: File/Directory locations to build your cache from
* `debug`: Enabling more logging for debugging
