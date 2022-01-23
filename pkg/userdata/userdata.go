package userdata

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/slim-ai/mob-code-server/pkg/config"
	"gopkg.in/yaml.v2"
)

func RunProvisioningScripts(ctx *pulumi.Context, settings *config.Settings,
	dependsOns []pulumi.Resource, templateFileHandler func(script string) string) error {
	if templateFileHandler == nil {
		return errors.New("templateFileHandler cannot be nil")
	}
	if scripts, err := getProvisioningScripts(settings.MachineInfo.OsDist); err != nil {
		return err
	} else {
		var instanceId pulumi.StringOutput
		// First extract the instance Id
		for _, dep := range dependsOns {
			switch inst := dep.(type) {
			case *ec2.SpotInstanceRequest:
				instanceId = pulumi.Sprintf("%s", inst.ID())
			default:
			}
		}
		for _, entry := range scripts {
			createName := filepath.Base(entry.Up)
			createScriptText, err := ioutil.ReadFile(entry.Up)
			if err != nil {
				return err
			}
			// XXX: This should be moved out to the settings
			defaultUser := "ubuntu"
			if settings.MachineInfo.OsDist == "arch" {
				defaultUser = "arch"
			}
			pulumi.Printf("Running provisioning script [%s]\n", createName)
			createResolvedText := templateFileHandler(string(createScriptText))
			//
			args := &remote.CommandArgs{
				Connection: remote.ConnectionArgs{
					Host:       pulumi.String(settings.DomainName),
					Port:       pulumi.Float64(22),
					PrivateKey: settings.MachineInfo.Credentials.PrivateOutput,
					User:       pulumi.String(defaultUser),
				},
				Create: pulumi.StringPtr(createResolvedText),
				Triggers: pulumi.Array{
					instanceId,
				},
			}
			// If there is something to when tearing down, add it
			if entry.Down != "" {
				deleteScriptText, err := ioutil.ReadFile(entry.Down)
				if err != nil {
					return err
				}
				deleteResolvedText := templateFileHandler(string(deleteScriptText))
				args.Delete = pulumi.StringPtr(deleteResolvedText)
			}
			// Run it
			if cmd, err := remote.NewCommand(ctx, createName, args,
				pulumi.DependsOn(dependsOns),
			); err != nil {
				pulumi.Printf("%s failed\n", createName)
				if cmd != nil {
					pulumi.Printf("standard out: %s\n", cmd.Stdout)
					pulumi.Printf("standard err: %s\n", cmd.Stderr)
				}
				return err
			} else {
				// Force in order execution
				dependsOns = append(dependsOns, cmd)
			}
		}
	}
	return nil
}

func BuildUserData(settings *config.Settings, templateFileHandler func(script string) string) (string, error) {
	if scripts, err := getUserDataScripts(settings.MachineInfo.OsDist); err != nil {
		return "", err
	} else {
		userDataParts := make([]string, len(scripts))
		for i, scriptFile := range scripts {
			if scriptText, err := ioutil.ReadFile(scriptFile); err != nil {
				return "", err
			} else {
				if templateFileHandler != nil {
					userDataParts[i] = templateFileHandler(string(scriptText))
				}
			}
		}
		return strings.Join(userDataParts, "\n###\n"), nil
	}
}

type ProvisioningSequence struct {
	Sequence []SeqEntry `yaml:"sequence"`
}

type SeqEntry struct {
	Up   string `yaml:"up"`
	Down string `yaml:"down"`
}

func getUserDataScripts(osDist string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// Relative directory path from project/cmd
	scriptDir := filepath.Clean(filepath.Join(cwd, fmt.Sprintf("../scripts/%s/userdata", osDist)))
	entries, err := getScripts(scriptDir)
	if err != nil {
		return nil, err
	}
	datascripts := make([]string, len(entries))
	for i, entry := range entries {
		if entry.Up != "" {
			datascripts[i] = entry.Up
		}
	}
	return datascripts, nil
}

func getProvisioningScripts(osDist string) ([]SeqEntry, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// Relative directory path from project/cmd
	scriptDir := filepath.Clean(filepath.Join(cwd, fmt.Sprintf("../scripts/%s/provisioning", osDist)))
	return getScripts(scriptDir)
}

func getScripts(directory string) ([]SeqEntry, error) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, err
	}
	orderFile := filepath.Join(directory, "sequence.yml")
	if _, err := os.Stat(orderFile); os.IsNotExist(err) {
		return nil, err
	}
	b, err := ioutil.ReadFile(orderFile)
	if err != nil {
		return nil, err
	}
	cfg := ProvisioningSequence{}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}
	files := make([]SeqEntry, len(cfg.Sequence))
	for i, entry := range cfg.Sequence {
		if entry.Up == "" {
			continue
		}
		item := SeqEntry{
			Up: filepath.Clean(filepath.Join(directory, entry.Up)),
		}
		if _, err := os.Stat(item.Up); os.IsNotExist(err) {
			return nil, err
		}
		if entry.Down != "" {
			item.Down = filepath.Clean(filepath.Join(directory, entry.Down))
			if _, err := os.Stat(item.Down); os.IsNotExist(err) {
				return nil, err
			}
		}
		files[i] = item
	}
	return files, nil
}
