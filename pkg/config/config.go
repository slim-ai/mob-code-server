package config

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Settings struct {
	DomainName     string                       `yaml:"_" json:"_"`         // computed
	Email          string                       `yaml:"email" json:"email"` // populated when we init the CVS certs
	HostedZone     string                       `yaml:"hosted_zone" json:"hosted_zone"`
	VpcId          string                       `yaml:"vpc_id" json:"vpc_id"`
	MachineInfo    MachineInfo                  `yaml:"instance" json:"instance"`
	Gitlab         ConcurrentVersionsSystemInfo `yaml:"gitlab" json:"gitlab"`
	Github         ConcurrentVersionsSystemInfo `yaml:"github" json:"github"`
	ExtraVariables map[string]string            `yaml:"variables" json:"variables"`
}

type MachineInfo struct {
	AmiId        string         `yaml:"_" json:"_"`
	SubnetId     string         `yaml:"_" json:"_"`
	Hostname     string         `yaml:"hostname" json:"hostname"`
	UserName     string         `yaml:"username" json:"username"`
	InstanceType string         `yaml:"instance_type" json:"instance_type"`
	SpotPrice    string         `yaml:"spot_price" json:"spot_price"`
	OsDist       string         `yaml:"os_dist" json:"os_dist"`
	DiskSizeGB   int            `yaml:"disk_size" json:"disk_size"`
	Credentials  SshCredentials `yaml:"credentials" json:"credentials"`
}

type ConcurrentVersionsSystemInfo struct {
	Enabled      bool     `yaml:"enabled" json:"enabled"`
	Token        string   `yaml:"token" json:"token"`
	Repositories []string `yaml:"repositories" json:"repositories"`
	Username     string   `yaml:"username" json:"username"`
}

type SshCredentials struct {
	Created bool   `yaml:"_" json:"_" `
	Public  string `yaml:"public" json:"public" `
	Private string `yaml:"private" json:"private"`
}

func (settings *Settings) Load(ctx *pulumi.Context) error {
	config.New(ctx, "").RequireObject("settings", settings)
	// Set some defaults and enforce mandatory config settings
	if settings.HostedZone == "" {
		return errors.New("hosted_zone must be set")
	}
	if settings.Email == "" {
		return errors.New("email must be set")
	}
	if settings.MachineInfo.Hostname == "" {
		return errors.New("instance.hostname must be set")
	}
	if !settings.Gitlab.Enabled && !settings.Github.Enabled {
		return errors.New("must enable either github or gitlab")
	}
	if settings.Gitlab.Enabled && settings.Gitlab.Token == "" {
		return errors.New("must provide a Gitlab token")
	}
	if settings.Gitlab.Enabled && settings.Gitlab.Username == "" {
		return errors.New("must provide a Gitlab username")
	}
	if settings.Github.Enabled && settings.Github.Token == "" {
		return errors.New("must provide a Github token")
	}
	if settings.Github.Enabled && settings.Github.Username == "" {
		return errors.New("must provide a Github username")
	}
	// Decode the supplied priv key
	if settings.MachineInfo.Credentials.Private != "" {
		sDec, e := base64.StdEncoding.DecodeString(settings.MachineInfo.Credentials.Private)
		if e != nil {
			return e
		}
		settings.MachineInfo.Credentials.Private = string(sDec)
	}
	// Set some defaults if not set
	if settings.MachineInfo.OsDist == "" {
		settings.MachineInfo.OsDist = "ubuntu"
	}
	if settings.MachineInfo.UserName == "" {
		settings.MachineInfo.UserName = "coder"
	}
	if settings.MachineInfo.InstanceType == "" {
		settings.MachineInfo.UserName = "t3.large"
	}
	if settings.MachineInfo.DiskSizeGB == 0 {
		settings.MachineInfo.DiskSizeGB = 128
	}
	if settings.ExtraVariables == nil {
		settings.ExtraVariables = map[string]string{}
	}
	settings.DomainName = fmt.Sprintf("%s.%s", settings.MachineInfo.Hostname, settings.HostedZone)
	return nil
}
