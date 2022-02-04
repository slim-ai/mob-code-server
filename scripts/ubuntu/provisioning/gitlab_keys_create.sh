#!/usr/bin/env bash

(
    cd ~___USERNAME___/

    # Generate an SSH key
    sudo -u ___USERNAME___ ssh-keygen -t ed25519  -f ~___USERNAME___/.ssh/id_ed25519 -q -N ""

    # Pull down gitadm helper
    sudo -u ___USERNAME___ /usr/local/go/bin/go install github.com/slim-ai/gitadm@latest

    ## Push the key to gitlab
    sudo -u ___USERNAME___ ~___USERNAME___/go/bin/gitadm --token="___GITLAB_TOKEN___" \
      add ssh-key --title "___DOMAIN_NAME___" --overwrite=true --file ~___USERNAME___/.ssh/id_ed25519.pub

    ## Setup private org
    GOPRIVATE_ORGS=$(sudo -u ___USERNAME___  ~___USERNAME___/go/bin/gitadm describe orgs --short)
    echo "export GOPRIVATE=${GOPRIVATE_ORGS}" | sudo -u ___USERNAME___ tee -a ~___USERNAME___/.bashrc

# Setup the git config
cat > ~/.gitconfig <<EOF
[user]
    email = ___EMAIL__ADDRESS___
# Enforce SSH
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
[url "ssh://git@gitlab.com/"]
    insteadOf = https://gitlab.com/
EOF

   cat ~/.gitconfig | sudo -u ___USERNAME___ tee  ~___USERNAME___/.gitconfig
)