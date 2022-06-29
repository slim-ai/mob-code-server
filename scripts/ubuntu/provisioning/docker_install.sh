#!/usr/bin/env bash

###############################################
# Docker
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo \
  "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update -y
sudo apt-get install -y docker-ce docker-ce-cli containerd.io

sudo usermod -aG docker ___USERNAME___

# compose
DOCKER_CONFIG="~___USERNAME___/.docker"
sudo mkdir -p $DOCKER_CONFIG/cli-plugins
sudo curl -SL https://github.com/docker/compose/releases/download/v2.6.0/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose
sudo chown -R ___USERNAME___:___USERNAME___ $DOCKER_CONFIG

sudo cp  $DOCKER_CONFIG/cli-plugins/docker-compose /usr/bin
sudo chmod a+rx /usr/bin/docker-compose
