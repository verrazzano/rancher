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
			&cli.StringFlag{
				Name: "oci-region",
			},
			&cli.StringFlag{
				Name: "oci-passphrase",
			},
		}
	}
	return flags, nil
}
