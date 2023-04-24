package watcher

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rancher/rancher/pkg/controllers/managementagent/workload"
)

type workloadFetcher struct {
	workloadController workload.CommonController
}

func (w *workloadFetcher) getWorkloadName(namespace, name, kind string) (string, error) {
	if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" || kind == "CronJob" {
		return name, nil
	}

	workloadID := fmt.Sprintf("%s:%s:%s", kind, namespace, name)
	workload, err := w.workloadController.GetByWorkloadID(workloadID)
	if err != nil {
		return "", errors.Wrapf(err, "get workload %s failed", workloadID)
	}

	allRef := workload.OwnerReferences
	if len(allRef) == 0 {
		return name, nil
	}

	ref := allRef[0]
	refName := ref.Name
	refKind := ref.Kind

	if kind == "Job" && refKind != "CronJob" {
		return name, nil
	}

	refWorkloadID := fmt.Sprintf("%s:%s:%s", refKind, namespace, refName)
	refWorkload, err := w.workloadController.GetByWorkloadID(refWorkloadID)
	if err != nil {
		return "", errors.Wrapf(err, "get workload %s failed", workloadID)
	}

	return w.getWorkloadName(refWorkload.Namespace, refWorkload.Name, refWorkload.Kind)
}
