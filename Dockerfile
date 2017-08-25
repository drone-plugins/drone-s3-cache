FROM centurylink/ca-certs
ENV GODEBUG=netdns=go

ADD contrib/mime.types /etc/
ADD release/linux/amd64/drone-s3-cache /bin/
ENTRYPOINT ["/bin/drone-s3-cache"]
