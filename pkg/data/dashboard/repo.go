package dashboard

import (
	"context"
	"github.com/rancher/rancher/pkg/features"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/rancher/pkg/settings"

	v1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	"github.com/rancher/rancher/pkg/wrangler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addRepo(wrangler *wrangler.Context, gitRepo, repoName, branchName string) error {
	repo, err := wrangler.Catalog.ClusterRepo().Get(repoName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = wrangler.Catalog.ClusterRepo().Create(&v1.ClusterRepo{
			ObjectMeta: metav1.ObjectMeta{
				Name: repoName,
			},
			Spec: v1.RepoSpec{
				GitRepo:   gitRepo,
				GitBranch: branchName,
			},
		})
	} else if err == nil && repo.Spec.GitBranch != branchName {
		repo.Spec.GitBranch = branchName
		_, err = wrangler.Catalog.ClusterRepo().Update(repo)
	}

	return err
}

func addRepos(ctx context.Context, wrangler *wrangler.Context) error {
	if err := addRepo(wrangler, "https://github.com/verrazzano/rancher-charts", "rancher-charts", settings.ChartDefaultBranch.Get()); err != nil {
		return err
	}
	if err := addRepo(wrangler, "https://git.rancher.io/partner-charts", "rancher-partner-charts", settings.PartnerChartDefaultBranch.Get()); err != nil {
		return err
	}

	if features.RKE2.Enabled() {
		if err := addRepo(wrangler, "https://git.rancher.io/rke2-charts", "rancher-rke2-charts", settings.RKE2ChartDefaultBranch.Get()); err != nil {
			return err
		}
	}

	return nil
}
