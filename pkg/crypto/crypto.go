package crypto

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/kevinburke/ssh_config"
	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/slim-ai/mob-code-server/pkg/config"
)

// TryCreateMachineSshCertificate - creates a Cert for SSH access of the new machine
// if the user didn't provide one in settings
func TryCreateMachineSshCertificate(ctx *pulumi.Context, settings *config.Settings) error {
	sshDirectory := path.Clean(path.Join(os.Getenv("HOME"), ".ssh"))
	if _, err := os.Stat(sshDirectory); !os.IsNotExist(err) {
		if err := os.MkdirAll(sshDirectory, 0700); err != nil {
			return err
		}
	}

	if settings.MachineInfo.Credentials.Public == "" && settings.MachineInfo.Credentials.Private == "" {
		key, err := tls.NewPrivateKey(ctx,
			settings.DomainName,
			&tls.PrivateKeyArgs{
				Algorithm: pulumi.String("RSA"),
				RsaBits:   pulumi.IntPtr(4096),
			},
		)
		if err != nil {
			return err
		}
		keyFileName := fmt.Sprintf("%s/%s", sshDirectory, settings.DomainName)
		if _, err := local.NewCommand(ctx,
			"private-ssh-key-writer",
			&local.CommandArgs{
				Create: pulumi.Sprintf("echo '%s' > %s && chmod 400 %s", key.PrivateKeyPem, keyFileName, keyFileName),
				Delete: pulumi.Sprintf("rm -f %s", keyFileName),
			},
			pulumi.DependsOn([]pulumi.Resource{key}),
		); err != nil {
			return err
		}
		if _, err := local.NewCommand(ctx,
			"public-ssh-key-writer",
			&local.CommandArgs{
				Create: pulumi.Sprintf("echo '%s' > %s.pub", key.PublicKeyOpenssh, keyFileName),
				Delete: pulumi.Sprintf("rm -f %s.pub", keyFileName),
			},
			pulumi.DependsOn([]pulumi.Resource{key}),
		); err != nil {
			return err
		}
		settings.MachineInfo.Credentials.PrivateOutput = pulumi.Sprintf("%s", key.PrivateKeyPem)
		settings.MachineInfo.Credentials.PublicOutput = pulumi.Sprintf("%s", key.PublicKeyOpenssh)
		settings.MachineInfo.Credentials.Created = true
	} else {
		privateKey := pulumi.String(settings.MachineInfo.Credentials.Private)
		settings.MachineInfo.Credentials.PrivateOutput = privateKey.ToStringOutput().ApplyT(func(s string) string {
			return s
		}).(pulumi.StringOutput)
		publicKey := pulumi.String(settings.MachineInfo.Credentials.Public)
		settings.MachineInfo.Credentials.PublicOutput = publicKey.ToStringOutput().ApplyT(func(s string) string {
			return s
		}).(pulumi.StringOutput)
	}
	if err := tryWriteSshConfigFile(settings.MachineInfo.UserName, sshDirectory, settings.DomainName); err != nil {
		return err
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
	cfgFilename := path.Join(sshDirectory, "config")
	f, err := os.OpenFile(
		cfgFilename,
		os.O_APPEND|os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return err
	}
	defer f.Close()
	// Check for existing record
	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return err
	}
	for _, host := range cfg.Hosts {
		for _, key := range host.Patterns {
			if key.String() == certFileName {
				// found
				return nil
			}
			// skip complex rules...
			// the one we are interested in only has one entry
			break
		}
	}
	fmt.Printf("adding %s to %s\n", certFileName, cfgFilename)
	// Create a new entry
	newConfigFile := strings.ReplaceAll(template, "__DOMAINNAME__", certFileName)
	newConfigFile = strings.ReplaceAll(newConfigFile, "__USERNAME__", "ubuntu")
	newConfigFile = strings.ReplaceAll(newConfigFile, "__CERTFILEPATH__", path.Join(sshDirectory, certFileName))
	if _, err = f.WriteString(newConfigFile); err != nil {
		return err
	}
	return nil
}
