// Copyright (c) 2023, Oracle and/or its affiliates.

// This file from the Rancher repository has been modified by Oracle as follows:
// - references to the etcdsnapshots.rke.cattle.io CRDs and APIs have been removed

package v1

type ETCDSnapshotS3 struct {
	Endpoint            string `json:"endpoint,omitempty"`
	EndpointCA          string `json:"endpointCA,omitempty"`
	SkipSSLVerify       bool   `json:"skipSSLVerify,omitempty"`
	Bucket              string `json:"bucket,omitempty"`
	Region              string `json:"region,omitempty"`
	CloudCredentialName string `json:"cloudCredentialName,omitempty"`
	Folder              string `json:"folder,omitempty"`
}

type ETCDSnapshotCreate struct {
	// Changing the Generation is the only thing required to initiate a snapshot creation.
	Generation int `json:"generation,omitempty"`
}

type ETCDSnapshotRestore struct {
	// Name refers to the name of the associated etcdsnapshot object
	Name string `json:"name,omitempty"`

	// Changing the Generation is the only thing required to initiate a snapshot restore.
	Generation int `json:"generation,omitempty"`
	// Set to either none (or empty string), all, or kubernetesVersion
	RestoreRKEConfig string `json:"restoreRKEConfig,omitempty"`
}

type ETCD struct {
	DisableSnapshots     bool            `json:"disableSnapshots,omitempty"`
	SnapshotScheduleCron string          `json:"snapshotScheduleCron,omitempty"`
	SnapshotRetention    int             `json:"snapshotRetention,omitempty"`
	S3                   *ETCDSnapshotS3 `json:"s3,omitempty"`
}
