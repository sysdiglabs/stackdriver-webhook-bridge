# Introduction

This contains a simple go program that can read [Stackdriver K8s Audit Logs](https://cloud.google.com/kubernetes-engine/docs/how-to/audit-logging) and forward them to a configurable webhook. This is a useful way to route K8s Audit Events to a program like [falco](https://github.com/falcosecurity/falco) or the [Sysdig](https://sysdig.com/) Agent.

# Installation

These instructions assume you already have created a cluster and have configured the `gcloud` and `kubectl` command line programs to interact with the cluster.

1. Create a google cloud (not k8s) service account and key that has the ability to read logs:

    ```
    $ gcloud iam service-accounts create swb-logs-reader --description "Service account used by stackdriver-webhook-bridge" --display-name "stackdriver-webhook-bridge logs reader"
    $ gcloud projects add-iam-policy-binding <your gce project id> --member serviceAccount:swb-logs-reader@<your gce project id>.iam.gserviceaccount.com --role 'roles/logging.viewer'
    $ gcloud iam service-accounts keys create $HOME/swb-logs-reader-key.json --iam-account swb-logs-reader@<your gce project id>.iam.gserviceaccount.com
    ```

1. Create a k8s secret containing the service account keys:

    ```
    kubectl create secret generic stackdriver-webhook-bridge --from-file=key.json=$HOME/swb-logs-reader-key.json
    ```

1. Deploy the bridge program to your cluster using the provided [stackdriver-webhook-bridge.yaml](./stackdriver-webhook-bridge.yaml) file:

    ```
    kubectl apply -f stackdriver-webhook-bridge.yaml -n sysdig-agent
    ```

The bridge program routes audit events to the domain name `sysdig-agent.sysdig-agent.svc.cluster.local`, which corresponds to the sysdig-agent service you created when you deployed the agent.

## Development

The [Makefile](./Makefile) has `binary`, `image`, and `test` targets. There are unit tests that test the converter, ensuring that log entries are converted to expected K8s Audit Events.

## Limitations

### Audit Events Do Not Contain requestObject

GKE Uses a K8s Audit Policy that emits a more limited set of information than the audit policy recommended by Sysdig. In particular, audit events for configmaps generally do not contain a `requestObject` field that contains the object that's being created/modified. As a result, the following Rules from [k8s_audit_rules.yaml](https://github.com/falcosecurity/falco/blob/dev/rules/k8s_audit_rules.yaml) will not trigger:
* Create/Modify Configmap With Private Credentials: The contents of configmaps are not included in audit logs, so the contents can not be examined for sensitive information.

### Pod Exec Does Not Include Command/Container

For many K8s distributions, an audit event representing a pod exec includes the command and specific container as arguments to the requestURI, for example:

```
"requestURI":"/api/v1/namespaces/default/pods/nginx-deployment-7998647bdf-phvq7/exec?command=bash&container=nginx1&container=nginx1&stdin=true&stdout=true&tty=true
```

In GKE, the audit event is missing those request parameters. This doesn't affect any of the Falco rules in [k8s_audit_rules.yaml](https://github.com/falcosecurity/falco/blob/dev/rules/k8s_audit_rules.yaml), but does limit the information that can be returned in the outputs of rules.

## See Also

If you would rather use google pub/sub to receive stackdriver logs and forward them to a webhook, check out https://github.com/codeonline-io/falco-gke-audit-bridge.
