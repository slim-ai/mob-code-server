package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"

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
		if err := writeKeyToFile(keyBytes, path.Join(sshDirectory, settings.DomainName)); err != nil {
			return err
		}
		if err := TryWriteSshConfigFile(settings.MachineInfo.UserName, sshDirectory, settings.DomainName); err != nil {
			return err
		}
	}
	return nil
}

// writePemToFile writes keys to a file
func writeKeyToFile(keyBytes []byte, saveFileTo string) error {
	if err := os.WriteFile(saveFileTo, keyBytes, 0600); err != nil {
		return err
	}
	return nil
}
