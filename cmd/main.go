package main

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/slim-ai/mob-code-server/pkg/config"
	"github.com/slim-ai/mob-code-server/pkg/crypto"
	"github.com/slim-ai/mob-code-server/pkg/server"
	"github.com/slim-ai/mob-code-server/pkg/userdata"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		settings := config.Settings{}
		if err := settings.Load(ctx); err != nil {
			return err
		}
		///////////////////////////////////////////////////////////////
		// Verify provided information
		//
		// hosted zone
		hostedZone, err := server.GetHostedZone(ctx, &settings)
		if err != nil {
			return err
		}
		//
		// Maybe create new ssh cert if the user didn't provide one in settings
		if err := crypto.TryCreateMachineSshCertificate(ctx, &settings); err != nil {
			return err
		}
		/////////////////////////////////////////////////////////////
		// Helper method to resolve variables from config
		// and from generated values
		variableResolver := func(script string) string {
			// Email address for FQDN certificate create/renewal
			script = strings.ReplaceAll(script, "___EMAIL__ADDRESS___", settings.Email)
			script = strings.ReplaceAll(script, "___USERNAME___", settings.MachineInfo.UserName)
			script = strings.ReplaceAll(script, "___HOSTNAME___", settings.MachineInfo.Hostname)
			script = strings.ReplaceAll(script, "___DOMAIN_NAME___", settings.DomainName)
			if settings.Gitlab.Enabled {
				// For preloading repositories from gitlab
				script = strings.ReplaceAll(script, "___GITLAB_TOKEN___", settings.Gitlab.Token)
				script = strings.ReplaceAll(script, "___GITLAB_REPOS___", strings.Join(settings.Gitlab.Repositories, ","))
			}
			if settings.Github.Enabled {
				// For preloading repositories from gitlab
				script = strings.ReplaceAll(script, "___GITHUB_TOKEN___", settings.Gitlab.Token)
				script = strings.ReplaceAll(script, "___GITHUB_REPOS___", strings.Join(settings.Gitlab.Repositories, ","))
			}
			for key, value := range settings.ExtraVariables { // replace any user provided
				script = strings.ReplaceAll(script, key, value)
			}
			return script
		}

		//
		////////////////////////////////////////////////////////////
		// Get current version of code server installation script
		userDataScript, err := userdata.BuildUserData(&settings, variableResolver)
		if err != nil {
			return err
		}
		//
		////////////////////////////////////////////////////////////
		// create the instance
		inst, err := server.CreateNewInstance(ctx, &settings, &userDataScript)
		if err != nil {
			return err
		}
		//
		// Map the Route 53 (DNS) record
		nm := fmt.Sprintf("%s-route", settings.DomainName)
		route, err := route53.NewRecord(ctx, nm,
			&route53.RecordArgs{
				ZoneId:  pulumi.String(hostedZone.Id),
				Name:    pulumi.String(settings.DomainName),
				Type:    pulumi.String("A"),
				Ttl:     pulumi.Int(300),
				Records: pulumi.StringArray{inst.PublicIp},
			},
		)
		if err != nil {
			return err
		}
		// Finally run any one shot provisioning
		if err := userdata.RunProvisioningScripts(ctx,
			&settings,
			[]pulumi.Resource{route, inst},
			variableResolver, // Seed w/ same variables
		); err != nil {
			return err
		}
		ctx.Export("dns_name", pulumi.String(settings.DomainName))
		return nil
	})
}
