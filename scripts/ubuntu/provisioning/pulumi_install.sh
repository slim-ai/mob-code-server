#!/usr/bin/env bash

###############################################
# Pulumi
curl -fsSL https://get.pulumi.com | sudo sh
sudo mv "$HOME/.pulumi" /home/___USERNAME___/
sudo chown -R ___USERNAME___:___USERNAME___  "/home/___USERNAME___/"
echo 'export PATH=/home/___USERNAME___/.pulumi/bin:$PATH' | sudo tee -a /home/___USERNAME___/.bashrc
