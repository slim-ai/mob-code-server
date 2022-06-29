#!/usr/bin/env bash

(
    cd ~___USERNAME___/
    echo -e "Host *\n\tStrictHostKeyChecking no" | sudo -u ___USERNAME___ tee -a ~___USERNAME___/.ssh/config
    sudo -u ___USERNAME___ mkdir -p code
    cd code
    echo "___GITLAB_REPOS___" | sudo -u ___USERNAME___ tee repo.list
    ## Now check out all the
    variable="___GITLAB_REPOS___"
    for i in $(echo $variable | sed "s/,/ /g")
    do
        sudo -u ___USERNAME___ git clone $i
    done
)