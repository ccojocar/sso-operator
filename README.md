# sso-operator

Single Sign-On Kubernetes [operator](https://coreos.com/operators/) for [dex](https://github.com/coreos/dex), which can provision, expose and manage a [SSO proxy](https://github.com/bitly/oauth2_proxy) for a Kubernetes service. 

## Installation

### Using Jenkins X

You can import this poject into your [Jenkins X](https://jenkins-x.io/) platform:

```
jx import --url https://github.com/jenkins-x/sso-operator.git
```

At this stage, Jenkins X will deploy automatically the opreator into your staging environment. After deployment, you can see the applications details with:

```
jx get apps
```

### Skaffold and Helm 

The operator can be also  installed using [skaffold](https://github.com/GoogleContainerTools/skaffold) and [helm](https://github.com/helm/helm) as follows:

```bash
export VERSION=latest
export DOCKER_REGISTRY=<YOUR DOCKER REGISTRY>
export KUBERNETES_NAMESPACE=<YOUR NAMESPACE>
make install-helm
```

