#!/usr/bin/env bash

(
    cd ~___USERNAME___/code
    ## Now check out all the
    variable="___GITLAB_REPOS___"
    for i in $(echo $variable | sed "s/,/ /g")
    do
        DIR=$(echo "${i}" | cut -d'/' -f2 | cut -d'.' -f1)
        (
            cd $DIR
            CHANGES=$(sudo -u ___USERNAME___ git status -s)
            if [ ! -z "${CHANGES}" ]; then
                sudo -u ___USERNAME___ git add .
                sudo -u ___USERNAME___ git commit -m "shutdown commit"
                sudo -u ___USERNAME___ git push
            fi
        )
    done
)