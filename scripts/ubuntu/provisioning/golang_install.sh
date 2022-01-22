#!/usr/bin/env bash
# shellcheck disable=SC2016
set -e

VERSION="1.17"

OS="$(uname -s)"
ARCH="$(uname -m)"

case $OS in
    "Linux")
        case $ARCH in
        "x86_64")
            ARCH=amd64
            ;;
        "aarch64")
            ARCH=arm64
            ;;
        "armv6" | "armv7l")
            ARCH=armv6l
            ;;
        "armv8")
            ARCH=arm64
            ;;
        .*386.*)
            ARCH=386
            ;;
        esac
        PLATFORM="linux-$ARCH"
    ;;
    "Darwin")
        PLATFORM="darwin-amd64"
    ;;
esac

if [ -z "$PLATFORM" ]; then
    echo "Your operating system is not supported by the script."
    exit 1
fi

## Hacks for installer
GOPATH="/home/___USERNAME___/go"
shell_profile="/home/___USERNAME___/.bashrc"

PACKAGE_NAME="go$VERSION.$PLATFORM.tar.gz"
TEMP_DIRECTORY=$(mktemp -d)

echo "Downloading $PACKAGE_NAME ..."
if hash wget 2>/dev/null; then
    wget -q https://storage.googleapis.com/golang/$PACKAGE_NAME -O "$TEMP_DIRECTORY/go.tar.gz"
else
    curl -s -o "$TEMP_DIRECTORY/go.tar.gz" https://storage.googleapis.com/golang/$PACKAGE_NAME
fi

if [ $? -ne 0 ]; then
    echo "Download failed! Exiting."
    exit 1
fi

echo "Extracting File..."
sudo mkdir -p "/home/___USERNAME___/go"
sudo mkdir -p "/usr/local/go/src"
sudo mkdir -p "/usr/local/go/bin"
sudo mkdir -p "/usr/local/go/pkg"
sudo chown -R ___USERNAME___:___USERNAME___  "/usr/local/go"

sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "$TEMP_DIRECTORY/go.tar.gz"
sudo chown -R ___USERNAME___:___USERNAME___  "/usr/local/go"

echo 'export GOROOT=/usr/local/go' | sudo tee -a /home/___USERNAME___/.bashrc
echo 'export GOPATH=/home/___USERNAME___/go' | sudo tee -a  /home/___USERNAME___/.bashrc
echo 'export PATH=$GOROOT/bin:$GOPATH/bin:$PATH' | sudo tee -a  /home/___USERNAME___/.bashrc
echo 'export GOPRIVATE="___GOPRIVATE___"' | sudo tee -a  /home/___USERNAME___/.bashrc

echo -e "\nGo $VERSION was installed into $GOROOT.\nMake sure to relogin into your shell or run:"
echo -e "\n\tsource $shell_profile\n\nto update your environment variables."
echo "Tip: Opening a new terminal window usually just works. :)"
sudo rm -f "$TEMP_DIRECTORY/go.tar.gz"
