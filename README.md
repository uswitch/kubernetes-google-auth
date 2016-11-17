# Kubernetes Authentication

This repository contains both a web application (assumed to be deployed to AppEngine) and a command-line tool to help authenticate with Kubernetes.

It currently assumes authentication against Google and that tokens will be issued to anyone authenticating from the specified domain (controlled through the `VALID_DOMAIN` environment variable).

## How it works

The web application is used by [Kubernetes Webhook provider](http://kubernetes.io/docs/admin/authentication/#webhook-token-authentication). The web application is used to both generate and validate bearer tokens.

Tokens are stored using Google Cloud Datastore and issued with a 12 hour expiry.

The command-line application makes it easier for command-line users to authenticate with Google and capture the bearer tokens issued by the web application. Upon authenticating the client will write config into `~/.kube/config` and set these credentials as the current context.

## Setting up

1. Deploy the web application. The project layout assumes deployment onto AppEngine but it should be easy to deploy anywhere (albeit you'll need credentials to talk to Cloud Datastore). For AppEngine it just means configuring an `app.yaml` with the environment variables:
  * `OAUTH2_CLIENT_ID`
  * `OAUTH2_CLIENT_SECRET`
  * `VALID_DOMAIN`
2. Build the command-line app `$ make`. This will install it into `$GOPATH/bin/kauth`.

## Authenticating

Run:

```
$ kauth --cluster=my-cluster-in-kubeconfig --auth-url=https://web.auth.app/
```

## Authors

* [Paul Ingles](https://github.com/pingles) / [@pingles](https://twitter.com/pingles).
* [Tom Taylor](https://github.com/t0mmyt)

This was hacked up in a few days as a proof-of-concept. It could definitely do with some love but "works for us" ;).