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
