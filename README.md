# sso-operator

Single Sign-On Kubernetes [operator](https://coreos.com/operators/) for [dex](https://github.com/coreos/dex), which can provision, expose and manage a [SSO proxy](https://github.com/bitly/oauth2_proxy) for a Kubernetes service. 

## Prerequisites

The operator requires a [dex](https://github.com/coreos/dex) identity provider, which could be installed into your cluster using this [helm chart](https://github.com/helm/charts/tree/master/stable/dex):

```
helm install --name dex --namesapce <DEX NAMESPACE> .
```

If you decide to install `dex` in a different `namespace` than the operator, you will have to enable in the operator helm chart, the job which installs the `gRPC` certificates.

To do this, open the `charts/sso-operator/values.yaml` file and update the following values:

```
dex.certs.install.create: true
dex.certs.install.sourceNamespace: <DEX NAMESPACE>
```
Also the `dex` service will have to be publicly exposed using an ingress controller of your choice.

## Installation

### Using Jenkins X

You can import this project into your [Jenkins X](https://jenkins-x.io/) platform:

```
jx import --url https://github.com/jenkins-x/sso-operator.git
```

At this stage, Jenkins X will deploy automatically the operator into your staging environment. After deployment, you can see the applications details with:

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

## Enable SSO

After installing the operator, you can enable Single Sign-On for a Kubernetes service by creating a SSO custom resource. 

Let's start by creating a basic Go http service with Jenkins X:

```
jx create quickstart -l Go --name golang-http
```

Within a few minutes, the service should be running in your staging environment. You can view the Kubernetes service created for it with:

```
kubectl get svc -n jx-staging

NAME           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)           AGE
golang-http    ClusterIP   10.15.250.117   <none>        80/TCP            1m
sso-operator   ClusterIP   10.15.244.220   <none>        80/TCP            6m
```

You can enable now the Single Sign-On for this service by creating a custom resource as follows:

```yaml
cat <<EOF | kubectl create -f -
apiVersion: "jenkins.io/v1"
kind: "SSO"
metadata:
  name: "sso-golang-http"
  namespace: jx-staging
spec:
  oidcIssuerUrl: "<YOUR DEX URL>"
  upstreamService: "golang-http"
  domain: "<YOUR DOMAIN>"
  tls: false
  proxyImage: "cosmincojocar/oauth2_proxy"
  proxyImageTag: "latest"
  proxyResources:
    limits:
      cpu: 100m
      memory: 256Mi
    requests:
      cpu: 80m
      memory: 128Mi
  cookieSpec:
    name: "sso-golang-http"
    expire: "168h"
    refresh: "60m"
    secure: false
    httpOnly: true
EOF
```

__Note:__ You will have to update *YOUR DEX URL* and *YOUR DOMAIN* with your specific values.

A SSO proxy will be automatically created by the operator and publicly exposed under your domain. You can see the proxy URL with:

```
kubectl get ingress -n jx-staging
NAME           HOSTS                                                                     ADDRESS          PORTS     AGE
golang-http    golang-http.jx-staging.35.187.37.181.nip.io                               35.240.115.197   80        1m
```

You can open now the URL in a browser and check if Single Sign-On works.
