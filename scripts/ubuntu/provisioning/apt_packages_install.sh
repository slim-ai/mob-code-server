#!/usr/bin/env bash

sudo apt-get update -y && sudo apt-get upgrade -y

###############################################
# Random packages
sudo apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    make \
    unzip
