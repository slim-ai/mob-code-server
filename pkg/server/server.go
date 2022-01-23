package server

import (
	"errors"
	"fmt"
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

//
// CreateSecurityGroup creates a security group for this machine
func CreateSecurityGroup(ctx *pulumi.Context, settings *config.Settings) (*ec2.SecurityGroup, error) {
	name := fmt.Sprintf("%s.sg", settings.DomainName)
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
						pulumi.String("0.0.0.0/0"),
					},
				},
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("HTTP traffic"),
					FromPort:    pulumi.Int(80),
					ToPort:      pulumi.Int(80),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
					},
				},
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("SSH traffic"),
					FromPort:    pulumi.Int(22),
					ToPort:      pulumi.Int(22),
					Protocol:    pulumi.String("tcp"),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"),
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
			PublicKey: settings.MachineInfo.Credentials.PublicOutput,
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

//
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

//
func selectArchAmiId(ctx *pulumi.Context) (*string, error) {
	return selectAmiId(ctx, []ec2.GetAmiIdsFilter{
		{Name: "name", Values: []string{"arch-linux-lts-hvm-*.x86_64-ebs"}},
	}, "093273469852")
}

//
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
	increaseOffer(settings)
	ctx.Export("price_offer", pulumi.String(fmt.Sprintf("%s/hr", settings.MachineInfo.SpotPrice)))
	return nil
}

func increaseOffer(settings *config.Settings) {
	if s, err := strconv.ParseFloat(settings.MachineInfo.SpotPrice, 32); err == nil {
		settings.MachineInfo.SpotPrice = fmt.Sprintf("%f", s+0.003)
	}
}

////////////////////////////////////////////

//
// Creates an instance provided the settings and userdata script
func CreateNewInstance(ctx *pulumi.Context, settings *config.Settings, userData *string) (*ec2.SpotInstanceRequest, error) {
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
	// TODO: Make this iterative
	return ec2.NewSpotInstanceRequest(
		ctx,
		settings.DomainName,
		&ec2.SpotInstanceRequestArgs{
			SpotPrice: pulumi.String(settings.MachineInfo.SpotPrice),
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
}

////////////////////////////////////////////////
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
