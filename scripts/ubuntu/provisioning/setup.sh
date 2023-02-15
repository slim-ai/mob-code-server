#!/usr/bin/env bash
#
# Each installation is a script function
# and the sequence is defined at the bottom of the file.

# add_user_to_docker_group "username"
add_user_to_docker_group() {
    local username=$1
    sudo usermod -aG docker $username
}

# install_aws_cli
install_aws_cli() {
    (
        cd /tmp
        sudo curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
        sudo unzip awscliv2.zip
        sudo ./aws/install
        # Clean up
        sudo rm -f awscliv2.zip
        sudo rm -rf aws
    )
}

# install_caddy "domain_name" "email_address" "username"
install_caddy() {
    local domain_name=$1
    local email_address=$2
    local username=$3
    # Install packages and add the caddy repository
    sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
    sudo apt update
    sudo apt install caddy

    # Obtain cert.
    sudo snap install core; sudo snap refresh core
    sudo snap install --classic certbot
    sudo ln -s /snap/bin/certbot /usr/bin/certbot
    sudo certbot certonly --noninteractive --agree-tos --no-eff-email --cert-name $domain_name --no-redirect -d  $domain_name -m $email_address --webroot -w /usr/share/caddy/

    echo "$domain_name" | sudo tee /etc/caddy/Caddyfile
    echo "tls $email_address" | sudo tee -a  /etc/caddy/Caddyfile
    echo "reverse_proxy 127.0.0.1:8080" | sudo tee -a  /etc/caddy/Caddyfile

    sudo systemctl reload caddy

    sudo systemctl restart code-server@$username
}

# install_code_server "username"
install_code_server() {
    local username=$1
    # install code-server service system-wide
    export HOME=/root
    curl -fsSL https://code-server.dev/install.sh | sudo sh

    # create a code-server user
    sudo adduser --disabled-password --gecos "" $username
    echo "$username ALL=(ALL:ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/$username
    sudo usermod -aG sudo $username

    # copy ssh keys from root
    sudo cp -r /root/.ssh /home/$username/.ssh
    sudo chown -R $username:$username /home/$username/.ssh

    # configure code-server to use --link with the "coder" user
    sudo mkdir -p /home/$username/.config/code-server

    PASSWD=$(date +%s | sha256sum | base64 | head -c 16 ; echo)
    CODESERVER_CONFIG="/home/$username/.config/code-server/config.yaml"
    echo "disable-telemetry: true" | sudo tee ${CODESERVER_CONFIG}
    echo "auth: password" | sudo tee -a ${CODESERVER_CONFIG}
    echo "password: ${PASSWD}" | sudo tee -a ${CODESERVER_CONFIG}

    sudo chown -R $username:$username /home/$username/.config

    # start and enable code-server and our helper service
    sudo systemctl enable --now code-server@$username
}


# install_docker_compose
install_docker_compose() {
    # Create the file
    cat > /tmp/docker-compose <<EOL
    #!/bin/bash
    docker compose \$@
EOL
    # Make the file executable
    chmod a+x /tmp/docker-compose
    sudo mv /tmp/docker-compose /usr/local/bin/docker-compose
}

# install_git_secret
install_git_secret() {
    (
        cd /tmp
        sudo git clone https://github.com/sobolevn/git-secret.git git-secret
        cd git-secret && sudo make build
        sudo PREFIX="/usr/local" make install
    )
}

# install_nvm "username" "0.39.3"
install_nvm() {
  local username=$1
  local nvm_version=$2

  # add user if it doesn't exist
  if ! id "$username" >/dev/null 2>&1; then
    useradd "$username"
  fi
  (
    cd /home/$username
    # install nvm as user
    sudo -u $username curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v${nvm_version}/install.sh | sudo -u $username bash
    sudo -u $username bash /home/$username/.nvm/nvm.sh    
    echo "nvm install v16.19.0" | sudo -u $username tee -a /home/$username/.bashrc
  )
}

# install_packages
install_packages(){
    # DOCKER
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo \
     "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
     $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list &> /dev/null
    
    # PACKER
    wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
    
    sudo apt-get update -y && sudo apt-get upgrade -y
    sudo snap install yq --channel=v3/stable
    
    # Install packages
    sudo apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release \
        make \
        git \
        gcc \
        unzip \
        docker-compose-plugin \
        jq \
        docker-ce docker-ce-cli containerd.io \
        default-jre openjdk-11-jdk-headless \
        packer

}

# install_regctl "username"
install_regctl() {
    username=$1
    sudo -u $username /usr/local/go/bin/go install github.com/regclient/regclient/cmd/regctl@latest@latest
}


# install_pulumi "username""
install_pulumi() {
    username=$1
    curl -fsSL https://get.pulumi.com | sudo -u $username sh
}

# install_serverless "2.64.1" "username"
install_serverless() {
    version=$1
    username=$2
    curl -o- -L https://slss.io/install | sudo -u $username VERSION=$version bash
}

install_session_manager_plugin() {
    # Download the session manager plugin
    curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb"

    # Install the plugin
    sudo dpkg -i session-manager-plugin.deb

    # Clean up the downloaded file
    rm session-manager-plugin.deb
}

# install_terraform_switcher
install_terraform_switcher() {
    curl -L https://raw.githubusercontent.com/warrensbox/terraform-switcher/release/install.sh | sudo bash
}

# setup_git_repos "username" "gitlab_repos"
setup_git_repos() {
    local username=$1
    local gitlab_repos=$2
    (
        # Change to the user's home directory
        cd /home/$username
        echo -e "Host *\n\tStrictHostKeyChecking no" | sudo -u $username tee -a /home/$username/.ssh/config
        sudo -u $username mkdir -p code
        cd code
        echo "$gitlab_repos" | sudo -u $username tee repo.list
        # Now check out all the
        variable="$gitlab_repos"
        for i in $(echo $variable | sed "s/,/ /g")
        do
            sudo -u $username git clone $i
        done
    )
}

# set_hostname "new_hostname" "username"
function set_hostname() {
    hostname=$1
    username=$2
    sudo hostnamectl set-hostname $hostname
    sudo systemctl restart code-server@$username
}


# add_git_ssh "username" "gitlab_token" "domain_name" "email_address"
add_git_ssh() {
    local username=$1
    local gitlab_token=$2
    local domain_name=$3
    local email_address=$4
    (
        # Change to the user's home directory
        cd /home/$username/
        sudo -u $username mkdir -p /home/$username/.ssh
        
        # Generate an SSH key
        sudo -u $username ssh-keygen -t ed25519  -f /home/$username/.ssh/id_ed25519 -q -N ""

        # Pull down gitadm helper
        sudo -u $username /usr/local/go/bin/go install github.com/slimdevl/gitadm@latest

        # Push the key to gitlab
        sudo -u $username /home/$username/go/bin/gitadm --token="$gitlab_token" \
        add ssh-key --title "$domain_name" --overwrite=true --file /home/$username/.ssh/id_ed25519.pub

        # Setup private org
        GOPRIVATE_ORGS=$(sudo -u $username /home/$username/go/bin/gitadm describe orgs --short)
        echo "GOPRIVATE=${GOPRIVATE_ORGS}" | sudo -u $username tee -a /home/$username/.bashrc

        # Setup the git config
        cat > /tmp/.gitconfig <<EOF
    [user]
        email = $email_address
    # Enforce SSH
    [url "ssh://git@github.com/"]
        insteadOf = https://github.com/
    [url "ssh://git@gitlab.com/"]
        insteadOf = https://gitlab.com/
EOF
        sudo chown $username:$username /tmp/.gitconfig
        sudo mv /tmp/.gitconfig /home/$username/.gitconfig
    )
}

# remove_ssh_key "username" "domain_name"
remove_ssh_key() {
    local username=$1
    local domain_name=$2
    sudo -u $username /home/$username/go/bin/gitadm rm ssh-key --title "$domain_name"
}

# golang "username" "1.17" "private.com"
install_go() {
    local username=$1
    local version=$2
    local go_private=$3
    local os
    local arch
    local platform
    local package_name
    local temp_directory
    local shell_profile="/home/$username/.bashrc"
    os="$(uname -s)"
    arch="$(uname -m)"

    case $os in
        "Linux")
            case $arch in
            "x86_64")
                arch=amd64
                ;;
            "aarch64")
                arch=arm64
                ;;
            "armv6" | "armv7l")
                arch=armv6l
                ;;
            "armv8")
                arch=arm64
                ;;
            .*386.*)
                arch=386
                ;;
            esac
            platform="linux-$arch"
        ;;
        "Darwin")
            platform="darwin-amd64"
        ;;
    esac

    if [ -z "$platform" ]; then
        echo "Your operating system is not supported by the script."
        exit 1
    fi

    package_name="go$version.$platform.tar.gz"
    temp_directory=$(mktemp -d)

    echo "Downloading $package_name ..."
    if hash wget 2>/dev/null; then
        wget -q https://storage.googleapis.com/golang/$package_name -O "$temp_directory/go.tar.gz"
    else
        curl -s -o "$temp_directory/go.tar.gz" https://storage.googleapis.com/golang/$package_name
    fi

    if [ $? -ne 0 ]; then
        echo "Download failed! Exiting."
        exit 1
    fi

    echo "Extracting File..."
    sudo -u $username mkdir -p "/home/$username/go"
    sudo mkdir -p "/usr/local/go/src"
    sudo mkdir -p "/usr/local/go/bin"
    sudo mkdir -p "/usr/local/go/pkg"
    sudo chmod -R a+rx "/usr/local/go"
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "$temp_directory/go.tar.gz"

    echo "export GOROOT=/usr/local/go" | sudo -u $username tee -a /home/$username/.bashrc
    echo "export GOPATH=/home/$username/go" | sudo -u $username tee -a  /home/$username/.bashrc
    echo "export PATH=\$GOROOT/bin:\$GOPATH/bin:\$PATH" | sudo -u $username tee -a  /home/$username/.bashrc

    echo -e "\nGo $version was installed into $GOROOT.\nMake sure to relogin into your shell or run:"
    echo -e "\n\tsource $shell_profile\n\nto update your environment variables."
    echo "Tip: Opening a new terminal window usually just works. :)"
    sudo rm -f "$temp_directory/go.tar.gz"
}

# install_aws_cli
set_def_vars() {
    local username=$1
    
    echo 'export PATH=${PATH}:/usr/local/bin' | sudo -u $username tee -a  /home/$username/.bashrc
    echo 'export PATH=${PATH}:${HOME}/bin' | sudo -u $username tee -a  /home/$username/.bashrc
    echo "export SAI_ENV_TYPE=local" | sudo -u $username tee -a  /home/$username/.bashrc
    echo "export SAI_ENV_NAME=local" | sudo -u $username tee -a  /home/$username/.bashrc
    echo "export SAI_ENV_ROLE=local" | sudo -u $username tee -a  /home/$username/.bashrc
}


# Installation Sequence
# These variables are replaced by the pulumi automation
# before writing the file to the remote machine then running it.
install_packages
# Creates User
install_code_server "___USERNAME___"
install_caddy "___DOMAIN_NAME___" "___EMAIL__ADDRESS___" "___USERNAME___"
add_user_to_docker_group "___USERNAME___"
install_docker_compose
install_go "___USERNAME___" "___GOLANG_VERSION___" "___GOPRIVATE___"
install_terraform_switcher
install_serverless "___SERVERLESS_VERSION___" "___USERNAME___"
install_nvm "___USERNAME___" "___NVM_VERSION___"
install_aws_cli
install_pulumi "___USERNAME___"
install_git_secret
set_hostname "___HOSTNAME___" "___USERNAME___"
add_git_ssh "___USERNAME___" "___GITLAB_TOKEN___" "___DOMAIN_NAME___" "___EMAIL__ADDRESS___"
setup_git_repos "___USERNAME___" "___GITLAB_REPOS___"
install_session_manager_plugin
install_regctl "___USERNAME___"
set_def_vars "___USERNAME___"
