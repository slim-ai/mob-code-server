#!/usr/bin/env bash

###############################################
# Serveless
curl -o- -L https://slss.io/install | sudo VERSION=2.64.1 bash
sudo mv "$HOME/.serverless" /home/___USERNAME___/
sudo chown -R ___USERNAME___:___USERNAME___  "/home/___USERNAME___/"
echo 'export PATH=/home/___USERNAME___/.serverless/bin:$PATH' | sudo tee -a /home/___USERNAME___/.bashrc
