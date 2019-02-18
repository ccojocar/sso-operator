# sso-operator

Single Sign-On Kubernetes [operator](https://coreos.com/operators/) for [dex](https://github.com/coreos/dex), which can provision, expose and manage a [SSO proxy](https://github.com/bitly/oauth2_proxy) for a Kubernetes service. 

## Architecture

![architecture](images/architecture.png?row=true)

## Installation 

### Using Jenkins X

You can install the operator and its dependencies with [Jenkins X](https://jenkins-x.io/). The only requirement is to have already allocated a DNS domain for your ingress controller.


You can execute the command bellow and then follow all the wizard steps:

```
jx create addon sso 
```

### Using Helm 

#### Prerequisites

The operator requires the [dex](https://github.com/coreos/dex) identity provider and the [cert-manager](https://github.com/jetstack/cert-manager) version `v.0.6.0` to be installed into your cluster. 
You can install `dex`using this [helm chart](https://github.com/jenkins-x/dex/tree/master/charts/dex), which pre-configures the `GitHub connector`, and relies on `cert-manager` 
to issue certificates for dex gRPC API.

Before starting the installation, you have to create a [GitHub OAuth App](https://github.com/settings/applications/new) which should have  as `callback` *https://DEX_DOMAIN/callback* URL.

You can install the chart as follows:
```
helm upgrade -i --namespace <NAMESAPCE> --wait --timeout 600 dex \
         --set domain="<DEX_DOMAIN>" \
         --set connectors.github.config.clientID="<CLIENT_ID>" \ 
         --set connectors.github.config.clientSecret="<CLIENT_SECRET>" \
         --set connectors.github.config.orgs={ORG1,ORG2} \
         .
```

The web endpoints provided by dex IdP have to be publicly exposed and secured with TLS. You can do  this pretty easy, if you have the [Jenkins X](https://jenkins-x.io/) installed into your cluster.

Just executing the command:

```
jx upgrade ingress 
```

You can select TLS and provide your `DEX_DOMAIN` and email. This command will configure the ingress controller to fetch automatically the TLS certificate from a Let's Encrypt CA server.

#### Install the operator

```
helm install --namespace <NAMESPACE> --set dex.grpcHost=dex.<DEX_NAMESPACE> charts/sso-operator/ 
```

## Enable Single Sign-On for a service 

After installing the operator, you can enable Single Sign-On for any Kubernetes service by creating a SSO custom resource. 

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
  oidcIssuerUrl: "https://dex.jx-staging.example.com"
  upstreamService: "golang-http"
  forwardToken: false
  domain: "example.com"
  certIssuerName: "letsencrypt-prod"
  urlTemplate: "{{.Service}}.{{.Namespace}}.{{.Domain}}"
  cookieSpec:
    name: "sso-golang-http"
    expire: "168h"
    refresh: "60m"
    secure: true
    httpOnly: true
  proxyImage: "quay.io/pusher/oauth2_proxy"
  proxyImageTag: "3.1.0"
  proxyResources:
    limits:
      cpu: 100m
      memory: 256Mi
    requests:
      cpu: 80m
      memory: 128Mi
EOF
```

__Note:__ You will have to update *oidcIssuerUrl* and *domain* with your specific values.

A SSO proxy will be automatically created by the operator and publicly exposed under your domain with TLS enabled. You can see the proxy URL with:

```
kubectl get ingress -n jx-staging
NAME              HOSTS                                                             ADDRESS        PORTS     AGE
sso-golang-http   sso-golang-http.jx-staging.example.com                            104.155.7.81   80, 443   37m
```

You can open now the `https://sso-golang-http.jx-staging.example.com` URL in a browser and check if Single Sign-On works with your GitHub user.
