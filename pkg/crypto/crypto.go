package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/slim-ai/mob-code-server/pkg/config"
	"golang.org/x/crypto/ssh"
)

// generatePrivateKey creates a RSA Private Key of specified byte size
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}
	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}
	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)
	return privatePEM
}

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func GeneratePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)
	return pubKeyBytes, nil
}

// TryCreateMachineSshCertificate - creates a Cert for SSH access of the new machine
// if the user didn't provide one in settings
func TryCreateMachineSshCertificate(settings *config.Settings) error {
	if settings.MachineInfo.Credentials.Public == "" {
		privKey, err := GeneratePrivateKey(2048)
		if err != nil {
			return err
		}
		pubKey, err := GeneratePublicKey(&privKey.PublicKey)
		if err != nil {
			return err
		}
		keyBytes := EncodePrivateKeyToPEM(privKey)
		settings.MachineInfo.Credentials.Private = string(keyBytes)
		settings.MachineInfo.Credentials.Public = string(pubKey)
		settings.MachineInfo.Credentials.Created = true

		//
		// Now write out to file
		sshDirectory := path.Clean(path.Join(os.Getenv("HOME"), ".ssh"))
		if _, err := os.Stat(sshDirectory); !os.IsNotExist(err) {
			if err := os.MkdirAll(sshDirectory, 0700); err != nil {
				return err
			}
		}
		if err := WriteKeyToFile(keyBytes, path.Join(sshDirectory, settings.DomainName)); err != nil {
			return err
		}
		if err := tryWriteSshConfigFile(settings.MachineInfo.UserName, sshDirectory, settings.DomainName); err != nil {
			return err
		}
	}
	return nil
}

const (
	template = `
Host __DOMAINNAME__
    Hostname __DOMAINNAME__
    User __USERNAME__
    Port 22
    IdentityFile __CERTFILEPATH__
`
)

// tryWriteSshConfigFile will try to create, or append to .ssh/config
// if the entry exist - no update is performed
func tryWriteSshConfigFile(username string, sshDirectory string, certFileName string) error {
	configFile := path.Join(sshDirectory, "config")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// it exists - check for existing entry
		b, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		if strings.Contains(string(b), certFileName) {
			return nil
		}
	}

	// Ok - write it
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	text := strings.ReplaceAll(template, "__DOMAINNAME__", certFileName)
	text = strings.ReplaceAll(text, "__USERNAME__", "ubuntu")
	text = strings.ReplaceAll(text, "__CERTFILEPATH__", path.Join(sshDirectory, certFileName))
	if _, err = f.WriteString(text); err != nil {
		return err
	}
	return nil
}

// writePemToFile writes keys to a file
func WriteKeyToFile(keyBytes []byte, saveFileTo string) error {
	err := ioutil.WriteFile(saveFileTo, keyBytes, 0600)
	if err != nil {
		return err
	}
	return nil
}
