#!/usr/bin/env bash

# install code-server service system-wide
export HOME=/root
curl -fsSL https://code-server.dev/install.sh | sudo sh

# create a code-server user
sudo adduser --disabled-password --gecos "" ___USERNAME___
echo "___USERNAME___ ALL=(ALL:ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/___USERNAME___
sudo usermod -aG sudo ___USERNAME___

# copy ssh keys from root
sudo cp -r /root/.ssh /home/___USERNAME___/.ssh
sudo chown -R ___USERNAME___:___USERNAME___ /home/___USERNAME___/.ssh

# configure code-server to use --link with the "coder" user
sudo mkdir -p /home/___USERNAME___/.config/code-server

PASSWD=$(date +%s | sha256sum | base64 | head -c 16 ; echo)
CODESERVER_CONFIG="/home/___USERNAME___/.config/code-server/config.yaml"
echo "disable-telemetry: true" | sudo tee ${CODESERVER_CONFIG}
echo "link: false" | sudo tee -a ${CODESERVER_CONFIG}
echo "auth: password" | sudo tee -a ${CODESERVER_CONFIG}
echo "password: ${PASSWD}" | sudo tee -a ${CODESERVER_CONFIG}

sudo chown -R ___USERNAME___:___USERNAME___ /home/___USERNAME___/.config

# start and enable code-server and our helper service
sudo systemctl enable --now code-server@___USERNAME___
