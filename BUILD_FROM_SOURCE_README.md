## Source Code Notes

### Pinning Bundled System Charts

The Rancher Dockerfile used to build the Rancher image uses `git clone` to embed Helm system charts in the image. The clone is done from a branch, but does not checkout a specific commit. There are commits to the Rancher partner chart repositories fairly regularly, which means every time we rebuild our forked Rancher image it could pull in a newer version of the system charts. This makes the Rancher image build nondeterministic with respect to the bundled system charts.

To fix this problem, we have modified the [Dockerfile](package/Dockerfile#L95-L102) to checkout specific commits in the system chart repositories. When building a new version of Rancher from source, you must determine the commit ids that are used in the upstream Rancher image and set them in the [Dockerfile](package/Dockerfile#L80-L82). The easiest way to determine the commit ids is to run the upstream Rancher image interactively and use `git`

For example, to find the relevant chart git commits for v2.6.8:
```
$ docker run --privileged -it --entrypoint=bash rancher/rancher:v2.6.8

b30c6a40e26c:/var/lib/rancher # git -C /var/lib/rancher-data/local-catalogs/v2/rancher-charts/4b40cac650031b74776e87c1a726b0484d0877c3ec137da0872547ff9b73a721 rev-parse HEAD
b08caead322a563a73b0eac24e9c14e77e3f8923

b30c6a40e26c:/var/lib/rancher # git -C /var/lib/rancher-data/local-catalogs/v2/rancher-partner-charts/8f17acdce9bffd6e05a58a3798840e408c4ea71783381ecd2e9af30baad65974 rev-parse HEAD
5bf8c3c4d76767610536051023fca99cf5bbaabf

b30c6a40e26c:/var/lib/rancher # git -C /var/lib/rancher-data/local-catalogs/v2/rancher-rke2-charts/675f1b63a0a83905972dcab2794479ed599a6f41b86cd6193d69472d0fa889c9 rev-parse HEAD
f08567172548acf8beffa4466b84580ba4220b03
```

## Build Instructions

The upstream tag this release is branched from is `v2.7.2-rc1`

### Create Environment Variables

```
export DOCKER_REPO=<Docker Repository>
export DOCKER_NAMESPACE=<Docker Namespace>
export DOCKER_TAG=<Docker Tag>
```

### Build and Push Images

By default, Rancher uses the latest tag on the Git branch as the image tag, so create the tag and run `make`:
```
git tag ${DOCKER_TAG}
make
```

Alternatively you can skip creating the tag and simply pass an environment variable to `make`:
```
TAG=${DOCKER_TAG} make
```

Once the build completes successfully, tag and push the images:
```
docker tag rancher/rancher:${DOCKER_TAG} ${DOCKER_REPO}/${DOCKER_NAMESPACE}/rancher:${DOCKER_TAG}
docker tag rancher/rancher-agent:${DOCKER_TAG} ${DOCKER_REPO}/${DOCKER_NAMESPACE}/rancher/rancher-agent:${DOCKER_TAG}
docker push ${DOCKER_REPO}/${DOCKER_NAMESPACE}/rancher:${DOCKER_TAG}
docker push ${DOCKER_REPO}/${DOCKER_NAMESPACE}/rancher/rancher-agent:${DOCKER_TAG}
```
