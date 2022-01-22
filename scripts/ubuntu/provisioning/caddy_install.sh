#!/usr/bin/env bash

sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/cfg/gpg/gpg.155B6D79CA56EA34.key' | sudo apt-key add -
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/cfg/setup/config.deb.txt?distro=debian&version=any-version' | sudo tee -a /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# Obtain cert.
sudo certbot-auto certonly

echo "___DOMAIN_NAME___" | sudo tee /etc/caddy/Caddyfile
echo "tls ___EMAIL__ADDRESS___" | sudo tee -a  /etc/caddy/Caddyfile
echo "reverse_proxy 127.0.0.1:8080" | sudo tee -a  /etc/caddy/Caddyfile

sudo systemctl reload caddy

sudo systemctl restart code-server@___USERNAME___
