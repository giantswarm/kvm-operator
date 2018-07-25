## Requirements

```
go get -u github.com/giantswarm/e2e-harness
```

## environment variables example
```
export CIRCLE_SHA1="master"
export CIRCLE_BUILD_NUM="990"
export CLUSTER_NAME="test-e2e"
export COMMON_DOMAIN="gastropod.gridscale.kvm.gigantic.io"
export IDRSA_PUB=$(cat ~/.ssh/id_rsa.pub)
export TESTED_VERSION="current"

export REGISTRY_PULL_SECRET=""
export GITHUB_BOT_TOKEN="xxxxxx"
```

## How to run integration test

```
$ minikube start
$ e2e-harness setup --remote=false
$ e2e-harness test --test-dir=integration/test/ready
```
