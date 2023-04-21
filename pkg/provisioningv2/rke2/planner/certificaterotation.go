package planner

import (
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	rkecontrollers "github.com/rancher/rancher/pkg/generated/controllers/rke.cattle.io/v1"
	"github.com/rancher/rancher/pkg/wrangler"
)

type certificateRotation struct {
	rkeControlPlanes rkecontrollers.RKEControlPlaneClient
	store            *PlanStore
}

func newCertificateRotation(clients *wrangler.Context, store *PlanStore) *certificateRotation {
	return &certificateRotation{
		rkeControlPlanes: clients.RKE.RKEControlPlane(),
		store:            store,
	}
}

const idempotentRotateScript = `
#!/bin/sh

currentGeneration=""
targetGeneration=$2
runtime=$1
shift
shift

dataRoot="/var/lib/rancher/$runtime/certificate_rotation"
generationFile="$dataRoot/generation"

currentGeneration=$(cat "$generationFile" || echo "")

if [ "$currentGeneration" != "$targetGeneration" ]; then
  $runtime certificate rotate  $@
else
	echo "certificates have already been rotated to the current generation."
fi

mkdir -p $dataRoot
echo $targetGeneration > "$generationFile"
`

// shouldRotateEntry returns true if the rotated services are applicable to the entry's roles.
func shouldRotateEntry(rotation *rkev1.RotateCertificates, entry *planEntry) bool {
	relevantServices := map[string]struct{}{}

	if len(rotation.Services) == 0 {
		return true
	}

	if isWorker(entry) {
		relevantServices["rke2-server"] = struct{}{}
		relevantServices["api-server"] = struct{}{}
		relevantServices["kubelet"] = struct{}{}
		relevantServices["kube-proxy"] = struct{}{}
		relevantServices["auth-proxy"] = struct{}{}
	}

	if isControlPlane(entry) {
		relevantServices["rke2-server"] = struct{}{}
		relevantServices["api-server"] = struct{}{}
		relevantServices["kubelet"] = struct{}{}
		relevantServices["kube-proxy"] = struct{}{}
		relevantServices["auth-proxy"] = struct{}{}
		relevantServices["controller-manager"] = struct{}{}
		relevantServices["scheduler"] = struct{}{}
		relevantServices["rke2-controller"] = struct{}{}
		relevantServices["admin"] = struct{}{}
		relevantServices["cloud-controller"] = struct{}{}
	}

	if isEtcd(entry) {
		relevantServices["etcd"] = struct{}{}
		relevantServices["kubelet"] = struct{}{}
		relevantServices["rke2-server"] = struct{}{}
	}

	for i := range rotation.Services {
		if _, ok := relevantServices[rotation.Services[i]]; ok {
			return true
		}
	}

	return false
}
