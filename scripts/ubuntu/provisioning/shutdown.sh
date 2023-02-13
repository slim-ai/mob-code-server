#!/usr/bin/env bash
#

# remove_ssh_key "username" "domain_name"
remove_ssh_key() {
    local username=$1
    local domain_name=$2
    sudo -u $username /home/$username/go/bin/gitadm rm ssh-key --title "$domain_name"
}

remove_ssh_key "___USERNAME___" "___DOMAIN_NAME___"