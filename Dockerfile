FROM ubuntu:xenial
#FROM golang:latest # --- Description:    Debian GNU/Linux 8.7 (jessie)

MAINTAINER A Wander <awander@gmail.com>

ENV DEBIAN_FRONTEND noninteractive

# Set the PORT environment variable
ENV PORT 8080
ENV PORT 7777

RUN \
  apt-get update && \
  apt-get -y install \
          software-properties-common \
          vim \
          pwgen \
          unzip \
          curl \
          make \
          wget \
          git-core && \
  rm -rf /var/lib/apt/lists/*

RUN \
	curl -O https://storage.googleapis.com/golang/go1.7.5.linux-amd64.tar.gz && \
	tar -C /usr/local -xzf go1.7.5.linux-amd64.tar.gz && \
	ln -s  /usr/local/go/bin/go /usr/bin/go && \
	ln -s  /usr/local/go/bin/gofmt /usr/bin/gofmt &&\
	ln -s  /usr/local/go/bin/godoc /usr/bin/godoc 

ENV GOPATH=/root/go
ENV PATH="/usr/local/go/bin:/root/go:${PATH}"

# Tell Docker what command to run when the container starts
CMD ["/bin/bash"]
