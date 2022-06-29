#!/usr/bin/env bash

sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# Obtain cert.
sudo snap install core; sudo snap refresh core
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
sudo certbot certonly --noninteractive --agree-tos --no-eff-email --cert-name ___DOMAIN_NAME___ --no-redirect -d  ___DOMAIN_NAME___ -m ___EMAIL__ADDRESS___ --webroot -w /usr/share/caddy/

echo "___DOMAIN_NAME___" | sudo tee /etc/caddy/Caddyfile
echo "tls ___EMAIL__ADDRESS___" | sudo tee -a  /etc/caddy/Caddyfile
echo "reverse_proxy 127.0.0.1:8080" | sudo tee -a  /etc/caddy/Caddyfile

sudo systemctl reload caddy

sudo systemctl restart code-server@___USERNAME___
