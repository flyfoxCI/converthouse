FROM golang:1.12.6 as builder

ARG RELEASE=dev

COPY . /go/src/github.com/deepfabric/converthouse
WORKDIR /go/src/github.com/deepfabric/converthouse

RUN make converthouse 'release_version=${RELEASE}'

FROM alpine:latest

COPY --from=builder /go/src/github.com/deepfabric/id/dist/converthouse /usr/local/bin/converthouse

RUN mkdir -p /var/converthouse/
RUN mkdir -p /var/lib/converthouse/

# Alpine Linux doesn't use pam, which means that there is no /etc/nsswitch.conf,
# but Golang relies on /etc/nsswitch.conf to check the order of DNS resolving
# (see https://github.com/golang/go/commit/9dee7771f561cf6aee081c0af6658cc81fac3918)
# To fix this we just create /etc/nsswitch.conf and add the following line:
RUN echo 'hosts: files mdns4_minimal [NOTFOUND=return] dns mdns4' >> /etc/nsswitch.conf

# Define default command.
ENTRYPOINT ["/usr/local/bin/converthouse"]
