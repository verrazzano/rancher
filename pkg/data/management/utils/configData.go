// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package utils

import (
	cli "github.com/rancher/machine/libmachine/mcnflag"
)

func GetCreateFlagsForDriver(driver string) ([]cli.Flag, error) {
	var flags []cli.Flag
	switch driver {
	case "google":
		flags = []cli.Flag{
			&cli.StringFlag{
				Name: "google-auth-encoded-json",
			},
		}

	case "amazonec2":
		flags = []cli.Flag{
			&cli.StringFlag{
				Name: "amazonec2-access-key",
			},
			&cli.StringFlag{
				Name: "amazonec2-secret-key",
			},
			&cli.StringFlag{
				Name: "amazonec2-session-token",
			},
			&cli.StringFlag{
				Name: "amazonec2-ami",
			},
			&cli.StringFlag{
				Name: "amazonec2-region",
			},
			&cli.StringFlag{
				Name: "amazonec2-vpc-id",
			},
			&cli.StringFlag{
				Name: "amazonec2-zone",
			},
			&cli.StringFlag{
				Name: "amazonec2-subnet-id",
			},
			&cli.BoolFlag{
				Name: "amazonec2-security-group-readonly",
			},
			&cli.StringSliceFlag{
				Name: "amazonec2-security-group",
			},
			&cli.StringSliceFlag{
				Name: "amazonec2-open-port",
			},
			&cli.StringFlag{
				Name: "amazonec2-tags",
			},
			&cli.StringFlag{
				Name: "amazonec2-instance-type",
			},
			&cli.StringFlag{
				Name: "amazonec2-device-name",
			},
			&cli.IntFlag{
				Name: "amazonec2-root-size",
			},
			&cli.StringFlag{
				Name: "amazonec2-volume-type",
			},
			&cli.StringFlag{
				Name: "amazonec2-iam-instance-profile",
			},
			&cli.StringFlag{
				Name: "amazonec2-ssh-user",
			},
			&cli.BoolFlag{
				Name: "amazonec2-request-spot-instance",
			},
			&cli.StringFlag{
				Name: "amazonec2-spot-price",
			},
			&cli.IntFlag{
				Name: "amazonec2-block-duration-minutes",
			},
			&cli.BoolFlag{
				Name: "amazonec2-private-address-only",
			},
			&cli.BoolFlag{
				Name: "amazonec2-use-private-address)",
			},
			&cli.BoolFlag{
				Name: "amazonec2-monitoring",
			},
			&cli.BoolFlag{
				Name: "amazonec2-use-ebs-optimized-instance",
			},
			&cli.StringFlag{
				Name: "amazonec2-ssh-keypath",
			},
			&cli.StringFlag{
				Name: "amazonec2-keypair-name",
			},
			&cli.IntFlag{
				Name: "amazonec2-retries",
			},
			&cli.StringFlag{
				Name: "amazonec2-endpoint",
			},
			&cli.BoolFlag{
				Name: "amazonec2-insecure-transport",
			},
			&cli.StringFlag{
				Name: "amazonec2-user-data",
			},
			&cli.BoolFlag{
				Name: "amazonec2-encrypt-ebs-volume",
			},
			&cli.StringFlag{
				Name: "amazonec2-kms-key",
			},
			&cli.StringFlag{
				Name: "amazonec2-http-endpoint",
			},
			&cli.StringFlag{
				Name: "amazonec2-http-tokens",
			},
		}

		/*
			(*mcnflag.StringFlag)(0xc0024a1600)(azure-environment),
			 (*mcnflag.StringFlag)(0xc0024a1640)(azure-subscription-id),
			 (*mcnflag.StringFlag)(0xc0024a1680)(azure-tenant-id),
			 (*mcnflag.StringFlag)(0xc0024a16c0)(azure-resource-group),
			 (*mcnflag.StringFlag)(0xc0024a1700)(azure-ssh-user),
			 (*mcnflag.IntFlag)(0xc0024a17c0)(azure-docker-port),
			 (*mcnflag.StringFlag)(0xc0024a1800)(azure-location),
			 (*mcnflag.StringFlag)(0xc0024a1840)(azure-size),
			 (*mcnflag.StringFlag)(0xc0024a1880)(azure-image),
			 (*mcnflag.StringFlag)(0xc0024a18c0)(azure-vnet),
			 (*mcnflag.StringFlag)(0xc0024a1900)(azure-subnet),
			 (*mcnflag.StringFlag)(0xc0024a1940)(azure-subnet-prefix),
			 (*mcnflag.StringFlag)(0xc0024a1980)(azure-availability-set),
			 (*mcnflag.StringFlag)(0xc0024a19c0)(azure-nsg),
			 (*mcnflag.StringFlag)(0xc0024a1a00)(azure-plan),
			 (*mcnflag.BoolFlag)(0xc002340660)(azure-managed-disks),
			 (*mcnflag.IntFlag)(0xc0024a1ac0)(azure-fault-domain-count),
			 (*mcnflag.IntFlag)(0xc0024a1b00)(azure-update-domain-count),
			 (*mcnflag.IntFlag)(0xc0024a1b40)(azure-disk-size),
			 (*mcnflag.StringFlag)(0xc0024a1b80)(azure-custom-data),
			 (*mcnflag.StringFlag)(0xc0024a1bc0)(azure-private-ip-address),
			 (*mcnflag.StringFlag)(0xc0024a1c00)(azure-storage-type),
			 (*mcnflag.BoolFlag)(0xc002340c60)(azure-use-private-ip),
			 (*mcnflag.BoolFlag)(0xc002340de0)(azure-no-public-ip),
			 (*mcnflag.BoolFlag)(0xc0023411a0)(azure-static-public-ip),
			 (*mcnflag.StringFlag)(0xc0024a1c40)(azure-dns),
			 (*mcnflag.StringSliceFlag)(0xc0084ec280)(azure-open-port),
			 (*mcnflag.StringFlag)(0xc0024a1d40)(azure-client-id),
			 (*mcnflag.StringFlag)(0xc0024a1d80)(azure-client-secret)
		*/
	case "azure":
		flags = []cli.Flag{
			&cli.StringFlag{
				Name: "azure-client-id",
			},
			&cli.StringFlag{
				Name: "azure-client-secret",
			},
			&cli.StringFlag{
				Name: "azure--environment",
			},
			&cli.StringFlag{
				Name: "azure-subscription-id",
			},
			&cli.StringFlag{
				Name: "azure-tenant-id",
			},
			&cli.StringFlag{
				Name: "azure-resource-group",
			},
			&cli.StringFlag{
				Name: "azure-ssh-user",
			},
			&cli.IntFlag{
				Name: "azure-docker-port",
			},
			&cli.StringFlag{
				Name: "azure-location",
			},
			&cli.StringFlag{
				Name: "azure-size",
			},
			&cli.StringFlag{
				Name: "azure-image",
			},
			&cli.StringFlag{
				Name: "azure-vnet",
			},
			&cli.StringFlag{
				Name: "azure-subnet",
			},
			&cli.StringFlag{
				Name: "azure-subnet-prefix",
			},
			&cli.StringFlag{
				Name: "azure-availability-set",
			},
			&cli.StringFlag{
				Name: "azure-nsg",
			},
			&cli.StringFlag{
				Name: "azure-plan",
			},
			&cli.BoolFlag{
				Name: "azure-managed-disks",
			},
			&cli.IntFlag{
				Name: "azure-fault-domain-count",
			},
			&cli.IntFlag{
				Name: "azure-update-domain-count",
			},
			&cli.IntFlag{
				Name: "azure-disk-size",
			},
			&cli.StringFlag{
				Name: "azure-custom-data",
			},
			&cli.StringFlag{
				Name: "azure-private-ip-address",
			},
			&cli.StringFlag{
				Name: "azure-storage-type",
			},
			&cli.BoolFlag{
				Name: "azure-use-private-ip",
			},
			&cli.BoolFlag{
				Name: "azure-no-private-ip",
			},
			&cli.BoolFlag{
				Name: "azure-static-private-ip",
			},
			&cli.StringFlag{
				Name: "azure-dns",
			},
			&cli.StringSliceFlag{
				Name: "azure-open-port",
			},
		}

	case "oci":
		flags = []cli.Flag{
			&cli.StringFlag{
				Name: "oci-fingerprint",
			},
			&cli.StringFlag{
				Name: "oci-private-key-contents",
			},
			&cli.StringFlag{
				Name: "oci-tenancy-id",
			},
			&cli.StringFlag{
				Name: "oci-user-id",
			},
		}
	}
	return flags, nil
}
