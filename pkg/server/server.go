package server

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/slim-ai/mob-code-server/pkg/config"
)

// GetVpcId returns the provided VpcId after validation,
// or if not provided (ie. ""), returns the default VpcId
func GetVpcId(ctx *pulumi.Context, settings *config.Settings) error {
	if settings.VpcId != "" {
		if _, err := ec2.LookupVpc(ctx,
			&ec2.LookupVpcArgs{Id: &settings.VpcId}, nil); err != nil {
			return err
		}
		return nil
	}
	opt := true
	// set to the default VPC
	vpcInfo, err := ec2.LookupVpc(ctx, &ec2.LookupVpcArgs{Default: &opt}, nil)
	if err != nil {
		return err
	}
	settings.VpcId = vpcInfo.Id
	return nil
}

////////////////////////////////////////////////////

// CreateSecurityGroup creates a security group for this machine
func CreateSecurityGroup(ctx *pulumi.Context, settings *config.Settings) (*ec2.SecurityGroup, error) {
	name := fmt.Sprintf("%s.sg", settings.DomainName)
	cidr := MyIpCidr()
	return ec2.NewSecurityGroup(
		ctx,
		name,
		&ec2.SecurityGroupArgs{
			Name:        pulumi.String(name),
			Description: pulumi.String(fmt.Sprintf("Security group for %s code server", settings.DomainName)),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("TLS traffic"),
					FromPort:    pulumi.Int(443),
					ToPort:      pulumi.Int(443),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"), // from anywhere for cert...
					},
				},
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("HTTP traffic"),
					FromPort:    pulumi.Int(80),
					ToPort:      pulumi.Int(80),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String(cidr),
					},
				},
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("SSH traffic"),
					FromPort:    pulumi.Int(22),
					ToPort:      pulumi.Int(22),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String(cidr),
					},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
			},
			Tags: pulumi.StringMap{
				"Owner": pulumi.String(settings.MachineInfo.Hostname),
			},
		},
	)
}

////////////////////////////////////////////

// CreateNewKeyPair records the key for the new instance
func CreateNewKeyPair(ctx *pulumi.Context, settings *config.Settings) (*ec2.KeyPair, error) {
	name := fmt.Sprintf("%s.kp", settings.DomainName)
	return ec2.NewKeyPair(ctx, name,
		&ec2.KeyPairArgs{
			KeyName:   pulumi.String(name),
			PublicKey: pulumi.String(settings.MachineInfo.Credentials.Public),
			Tags: pulumi.StringMap{
				"Owner": pulumi.String(settings.MachineInfo.Hostname),
			},
		},
	)
}

////////////////////////////////////////////

var (
	ErrUnsupportedDistribution error = errors.New("unsupported linux distribution")
	ErrNotFound                error = errors.New("unable to locate an AMI image")
)

// GetAmiId returns an AMI ID for the configured OS distribution
func GetAmiId(ctx *pulumi.Context, settings *config.Settings) error {
	type AmiSelector func(ctx *pulumi.Context) (*string, error)
	var selector AmiSelector
	switch strings.ToLower(settings.MachineInfo.OsDist) {
	case "ubuntu":
		selector = selectUbuntuAmiId
	case "arch":
		selector = selectArchAmiId
	default:
		return ErrUnsupportedDistribution
	}
	id, err := selector(ctx)
	if err != nil {
		return err
	}
	settings.MachineInfo.AmiId = *id
	return nil
}

func selectUbuntuAmiId(ctx *pulumi.Context) (*string, error) {
	return selectAmiId(ctx, []ec2.GetAmiIdsFilter{
		{Name: "architecture", Values: []string{"x86_64"}},
		{Name: "description", Values: []string{"*LTS*"}},
		{
			Name:   "name",
			Values: []string{"ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"},
		},
	}, "099720109477")
}

func selectArchAmiId(ctx *pulumi.Context) (*string, error) {
	return selectAmiId(ctx, []ec2.GetAmiIdsFilter{
		{Name: "name", Values: []string{"arch-linux-lts-hvm-*.x86_64-ebs"}},
	}, "093273469852")
}

func selectAmiId(ctx *pulumi.Context, filters []ec2.GetAmiIdsFilter, ownerId string) (*string, error) {
	amis, err := ec2.GetAmiIds(ctx, &ec2.GetAmiIdsArgs{
		ExecutableUsers: []string{},
		Filters:         filters,
		Owners:          []string{ownerId},
	}, nil)
	if err != nil {
		return nil, err
	}
	var selected *ec2.LookupAmiResult
	var selectedDate time.Time
	for _, ami := range amis.Ids {
		if info, err := ec2.LookupAmi(ctx,
			&ec2.LookupAmiArgs{
				Filters: []ec2.GetAmiFilter{{Name: "image-id", Values: []string{ami}}},
				Owners:  []string{ownerId},
			},
		); err != nil {
			return nil, err
		} else if selected != nil {
			infoDt, _ := time.Parse(time.RFC3339, info.CreationDate)
			if infoDt.Unix() > selectedDate.Unix() {
				selected = info
				selectedDate = infoDt
			}
		} else {
			selected = info
			infoDt, _ := time.Parse(time.RFC3339, info.CreationDate)
			selectedDate = infoDt
		}
	}
	if selected != nil {
		return &selected.Id, nil
	}
	return nil, ErrNotFound
}

////////////////////////////////////////////

func GetVpcIdPublicSubnet(ctx *pulumi.Context, settings *config.Settings) error {
	if subnets, err := ec2.GetSubnetIds(ctx, &ec2.GetSubnetIdsArgs{
		Filters: []ec2.GetSubnetIdsFilter{
			{Name: "vpc-id", Values: []string{settings.VpcId}},
			{Name: "state", Values: []string{"available"}},
			{Name: "default-for-az", Values: []string{"true"}},
		},
		VpcId: settings.VpcId,
	}); err != nil {
		return err
	} else if len(subnets.Ids) == 0 {
		return errors.New("unable to find subnet placement")
	} else {
		settings.MachineInfo.SubnetId = subnets.Ids[0]
	}
	return nil
}

func ValidateInstanceType(ctx *pulumi.Context, settings *config.Settings) error {
	if settings.MachineInfo.InstanceType == "" {
		settings.MachineInfo.InstanceType = "t3.large"
	}

	if _, err := ec2.GetInstanceType(ctx,
		&ec2.GetInstanceTypeArgs{
			InstanceType: settings.MachineInfo.InstanceType,
		}, nil); err != nil {
		return err
	}
	opt0 := "available"
	available, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{State: &opt0}, nil)
	if err != nil {
		return err
	}
	opt1 := "us-west-2a"
	if len(available.Names) > 0 {
		opt1 = available.Names[0]
	}
	priceInfo, err := ec2.GetSpotPrice(ctx, &ec2.GetSpotPriceArgs{
		AvailabilityZone: &opt1,
		Filters:          []ec2.GetSpotPriceFilter{{Name: "product-description", Values: []string{"Linux/UNIX"}}},
		InstanceType:     &settings.MachineInfo.InstanceType,
	}, nil)
	if err != nil {
		return err
	}
	settings.MachineInfo.SpotPrice = priceInfo.SpotPrice
	printPricing(ctx, settings)
	return nil
}

func printPricing(ctx *pulumi.Context, settings *config.Settings) {
	spotPricePerHour, _ := strconv.ParseFloat(settings.MachineInfo.SpotPrice, 64)
	if settings.MachineInfo.ResourceType == "ec2" {
		switch settings.MachineInfo.InstanceType {
		case "t3.nano":
			spotPricePerHour = 0.0052
		case "t3.micro":
			spotPricePerHour = 0.0104
		case "t3.small":
			spotPricePerHour = 0.0208
		case "t3.medium":
			spotPricePerHour = 0.0416
		case "t3.large":
			spotPricePerHour = 0.0832
		case "t3.xlarge":
			spotPricePerHour = 0.1664
		case "t3.2xlarge":
			spotPricePerHour = 0.332
		case "t3a.nano":
			spotPricePerHour = 0.0047
		case "t3a.micro":
			spotPricePerHour = 0.0094
		case "t3a.small":
			spotPricePerHour = 0.0188
		case "t3a.medium":
			spotPricePerHour = 0.0376
		case "t3a.large":
			spotPricePerHour = 0.0752
		case "t3a.xlarge":
			spotPricePerHour = 0.1504
		case "t3a.2xlarge":
			spotPricePerHour = 0.3008
		default:
			spotPricePerHour = 0.3008
		}
	}
	diskCostPerGbMo := 0.10 * float64(settings.MachineInfo.DiskSizeGB)
	diskPricePerHour := (diskCostPerGbMo * 12) / 8765.813
	totalCostPerHour := diskPricePerHour + spotPricePerHour
	if settings.MachineInfo.ResourceType == "spot" {
		ctx.Export("spot_price.market_price", pulumi.String(fmt.Sprintf("%.2f/hr USD", spotPricePerHour)))
		ctx.Export("spot_price.maximum_offer", pulumi.String(fmt.Sprintf("%s/hr USD", settings.MachineInfo.OfferSpotPrice)))
	} else {
		ctx.Export("ec2.price_guess", pulumi.String(fmt.Sprintf("%.2f/hr USD", spotPricePerHour)))
		ctx.Export("ec2.market_price", pulumi.String(fmt.Sprintf("see %s for cost/hr", "https://aws.amazon.com/ec2/pricing/on-demand/")))
	}
	ctx.Export("ebs.disk_price", pulumi.Sprintf("%f/mo USD", diskCostPerGbMo))
	ctx.Export("ebs.disk_size", pulumi.Sprintf("%d", settings.MachineInfo.DiskSizeGB))
	ctx.Export("estimate_total_cost.hr", pulumi.Sprintf("%f USD", totalCostPerHour))
	ctx.Export("estimate_total_cost.day", pulumi.Sprintf("%f USD", totalCostPerHour*24))
	ctx.Export("estimate_network_costs", pulumi.Sprintf("unknown"))
}

////////////////////////////////////////////

// Creates an instance provided the settings and userdata script
func CreateNewInstance(ctx *pulumi.Context, settings *config.Settings, hostedZone *route53.LookupZoneResult, userData *string) (pulumi.Resource, error) {
	if err := ValidateInstanceType(ctx, settings); err != nil {
		return nil, err
	}
	if err := GetAmiId(ctx, settings); err != nil {
		return nil, err
	}
	if err := GetVpcId(ctx, settings); err != nil {
		return nil, err
	}
	if err := GetVpcIdPublicSubnet(ctx, settings); err != nil {
		return nil, err
	}
	key, err := CreateNewKeyPair(ctx, settings)
	if err != nil {
		return nil, err
	}
	group, err := CreateSecurityGroup(ctx, settings)
	if err != nil {
		return nil, err
	}
	userDataScript := ""
	if userData != nil && len(*userData) > 0 {
		userDataScript = *userData
	}

	var (
		resource pulumi.Resource
		publicIp *pulumi.StringOutput
	)
	if settings.MachineInfo.ResourceType == "spot" {
		inst, err := ec2.NewSpotInstanceRequest(
			ctx,
			settings.DomainName,
			&ec2.SpotInstanceRequestArgs{
				SpotPrice: pulumi.String(settings.MachineInfo.OfferSpotPrice),
				Ami:       pulumi.String(settings.MachineInfo.AmiId),
				RootBlockDevice: ec2.SpotInstanceRequestRootBlockDeviceArgs{
					DeleteOnTermination: pulumi.Bool(true),
					VolumeSize:          pulumi.Int(settings.MachineInfo.DiskSizeGB),
					VolumeType:          pulumi.String("gp3"),
				},
				KeyName:             key.KeyName,
				InstanceType:        pulumi.String(settings.MachineInfo.InstanceType),
				UserData:            pulumi.String(userDataScript),
				VpcSecurityGroupIds: pulumi.StringArray{group.ID()},
				Tags: pulumi.StringMap{
					"Name":  pulumi.String(settings.DomainName),
					"Owner": pulumi.String(settings.MachineInfo.UserName),
				},
				WaitForFulfillment: pulumi.Bool(true),
			},
		)
		if err != nil {
			return nil, err
		}
		resource = inst
		publicIp = &inst.PublicIp
	} else {
		inst, err := ec2.NewInstance(ctx, settings.DomainName, &ec2.InstanceArgs{
			Ami:          pulumi.String(settings.MachineInfo.AmiId),
			InstanceType: pulumi.String(settings.MachineInfo.InstanceType),
			KeyName:      key.KeyName,
			RootBlockDevice: ec2.InstanceRootBlockDeviceArgs{
				DeleteOnTermination: pulumi.Bool(true), VolumeSize: pulumi.Int(settings.MachineInfo.DiskSizeGB), VolumeType: pulumi.String("gp3"),
			},
			Tags:                pulumi.StringMap{"Name": pulumi.String(settings.DomainName), "Owner": pulumi.String(settings.MachineInfo.UserName)},
			UserData:            pulumi.String(userDataScript),
			VpcSecurityGroupIds: pulumi.StringArray{group.ID()},
		})
		if err != nil {
			return nil, err
		}
		resource = inst
		publicIp = &inst.PublicIp
	}

	//
	// finally map the Route 53 (DNS) record
	nm := fmt.Sprintf("%s-route", settings.DomainName)
	route, err := route53.NewRecord(ctx, nm,
		&route53.RecordArgs{
			ZoneId:  pulumi.String(hostedZone.Id),
			Name:    pulumi.String(settings.DomainName),
			Type:    pulumi.String("A"),
			Ttl:     pulumi.Int(300),
			Records: pulumi.StringArray{*publicIp},
		},
		pulumi.DependsOn([]pulumi.Resource{resource}),
	)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func GetHostedZone(ctx *pulumi.Context, settings *config.Settings) (*route53.LookupZoneResult, error) {
	opt := false
	hostedZoneName := fmt.Sprintf("%s.", settings.HostedZone)
	if selected, err := route53.LookupZone(ctx,
		&route53.LookupZoneArgs{
			Name:        &hostedZoneName,
			PrivateZone: &opt,
		}, nil,
	); err != nil {
		return nil, err
	} else if selected == nil {
		return nil, errors.New("unable to locate hosted zone")
	} else {
		fmt.Printf("Found HostedZone.Id = %s\n", selected.ZoneId)
		return selected, nil
	}
}

func MyIpCidr() string {
	// Simplest way possible for IP address of this machine
	cmd := exec.Command("curl", "ifconfig.me")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "0.0.0.0/0"
	}
	return fmt.Sprintf("%s/32", strings.TrimSpace(out.String()))
}
