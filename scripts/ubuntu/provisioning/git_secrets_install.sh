#!/usr/bin/env bash

###############################################
# Git-secret
(
    cd /tmp
    sudo git clone https://github.com/sobolevn/git-secret.git git-secret
    cd git-secret && sudo make build
    sudo PREFIX="/usr/local" make install
)
