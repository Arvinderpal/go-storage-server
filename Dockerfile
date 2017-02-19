FROM ubuntu:xenial
MAINTAINER A Wander <awander@gmail.com>

ENV DEBIAN_FRONTEND noninteractive

RUN locale-gen en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

RUN \
  apt-get update && \
  apt-get -y install \
          software-properties-common \
          vim \
          pwgen \
          unzip \
          curl \
          git-core && \
  rm -rf /var/lib/apt/lists/*

CMD ["/bin/bash"]