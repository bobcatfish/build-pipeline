# Demo yaml

This folder contains `yaml` CRDs which can be used to demonstrate the current state of the
art in `Pipeline` functionality.

Features we can demonstrate:

1. [Decoupling](#decoupling)
2. [Typing](#typing)
3. [Cloud native](#cloud-native)

TODO:

* Make clusters actually work (mount a volume containing the kubeconfig so helm can use it? -
  this should be part of  #63, left a comment on https://github.com/knative/build-pipeline/pull/160)
* Output linking is broken (name of resource currently must have same name as Task uses, doesn't actually respect linking in Pipeline)
* Service account + Build + Kaniko isn't quite working, mounting into a volume instead

For the things I did start fixing, still need to:

* Fix tests
* Update examples

TODO later:

* If there is an error, the status of the TaskRun + PipelineRun doesn't get updated
* Reconciling never stops

## Setup

Before running any of these examples, you will need to
[setup the DEVELOPMENT requirements](https://github.com/knative/build-pipeline/blob/master/DEVELOPMENT.md#getting-started).

1. Create [two kubernetes clusters](https://github.com/knative/build-pipeline/blob/master/DEVELOPMENT.md#kubernetes-cluster)
   for yourself (one will be our `prod`, one will be `qa`).
   TODO: at the moment the `qa` cluster _needs_ to be the one that the Pipeline CRD is deployed to.
2. [Create a service account that can push to your registry](#service-account)
2. Replace the values in [`pipelineparams-qa.yaml`](pipelineparams-qa.yaml) and
   [`pipelineparams-prod.yaml`] with:
   1. Your cluster endpoint
   2. Your service account
3. Replace the `value`s in [`image.yaml`](image.yaml) with paths to images at a GCR registry
   which your service account can push to.

Deploy the Pipelines and Tasks to your cluster, note we won't be changing these, instead we'll
be reusing them with different parameters and Resources:

TODO: this is list is incomplete, I spent all my time getting `pipelinerun-qa.yaml` to work,
this is the list required for that:
```bash
kubectl apply -f image.yaml
kubectl apply -f kaniko.yaml
kubectl apply -f pipeline.yaml

kubectl apply -f skaffold-resource.yaml
kubectl apply -f pipelineparams-qa.yaml
kubectl apply -f pipelineparams-prod.yaml
```

### Service account

To [create a service account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
that can push to a GCR registry:

```bash
PROJECT_ID=your-gcp-project
ACCOUNT_NAME=scaffold-account
gcloud config set project $PROJECT_ID

# create the service account
gcloud iam service-accounts create $ACCOUNT_NAME --display-name $ACCOUNT_NAME
EMAIL=$(gcloud iam service-accounts list | grep $ACCOUNT_NAME | awk '{print $2}')

# add the storage.admin policy to the account so it can push containers
gcloud projects add-iam-policy-binding $PROJECT_ID --member serviceAccount:$EMAIL --role roles/storage.admin

# download the creds
gcloud iam service-accounts keys create config.json --iam-account $EMAIL

# create the secret and service account in your kubernetes cluster (do this in both clusters)
kubectl create secret generic skaffold-key --from-file=config.json
kubectl create -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: skaffold-account
secrets:
- name: skaffold-key
EOF

```

### Seeing logs

This is currently WIP, as of [#167](https://github.com/knative/build-pipeline/pull/167) we'll be
streaming the logs to a [`PersistentVolumeClaim`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/),
and to get these logs you would need to deploy another pod that reads from this volume.

## Decoupling

### Target environments

TODO: need to add support for clusters https://docs.helm.sh/helm/#options-inherited-from-parent-commands

`Pipelines` are decoupled from the environments they are run against, for example you can take the
`Pipeline` `deploy-pipeline.yaml` and run it with two separate sets of `PipelineParams`, one of which
will deploy to one environment, and the other will deploy to another environment.

1. Run [`pipeline.yaml`](pipeline.yaml) against your "qa" env:

   ```bash
   # This will work
   kubectl apply -f pipelinerun-qa.yaml
   ```

2. Run [`pipeline.yaml`](pipeline.yaml) against your "prod" env:

   ```bash
   # This will not work since we don't support clusters yet
   kubectl apply -f pipelinerun-qa.yaml
   ```

_Note that `PipelineRuns` must have unique names, so to re-run this you'll need to manually
change the `Name` field._

### Tasks from Pipelines

`Tasks` that are inside `Pipelines` are decoupled from the environments they run against,
and can be run without `Pipelines`.

To run the `Task` that builds images on its own:

1. Update [`buildimage-taskrun.yaml`](buildimage-taskrun.yaml) to use one of your kubernetes
   clusters and your service account.
2. Run the `TaskRun`:

   ```bash
   TODO
   taskrun.yaml
   ```

_Note that `TaskRuns` must have unique names, so to re-run this you'll need to manually
change the `Name` field._

### Github Resources

The revision `Resource` that a `Pipeline` runs against is not coupled to it, so you can
easily run a `Pipeline` against your own forks/branches.

1. Create a fork of [Skaffold](https://github.com/GoogleContainerTools/skaffold)
2. Update [`skaffoldfork-resource.yaml`](skaffold-fork-resource.yaml) to point at your
   fork.
3. Note that [`pipelinerun-fork.yaml`](pipelinerun-fork.yaml) uses the same
   [`pipeline.yaml`](pipeline.yaml) as [`pipelinerun-qa.yaml`](pipelinerun-qa.yaml)
   and [`pipelinerun-prod.yaml`](pipelinerun-prod.yaml), but it refers to the fork resource.
   Run it with:

   ```bash
   TODO
   kubectl apply -f skaffoldfork-resource.yaml
   kubectl apply -f pipelinerun-fork.yaml
   ```

_Note that `PipelineRuns` must have unique names, so to re-run this you'll need to manually
change the `Name` field._

## Typing

One of the features of the Pipeline CRD is that we have types associated with `Resources`
that are common for CI/CD pipelines that use k8s, for example you can use different
`Task` definitions to produce `Images` without having to make significant changes.

1. Run [`pipeline.yaml`](pipeline.yaml) against your "qa" env:

   ```bash
   kubectl apply -f pipelinerun-qa.yaml
   ```

2. Note that the `Pipeline` defined in [`pipeline-buildkit.yaml`](pipeline-buildkit.yaml) is the same
   as [`pipeline.yaml`](pipeline.yaml), however instead of referencing
   [the Task which builds with Kaniko](kaniko.yaml), it uses
   [the Task which builds with BuildKit](buildkit.yaml). Run it against your "qa" env:

   ```bash
   TODO
   ```

_Note that `PipelineRuns` must have unique names, so to re-run this you'll need to manually
change the `Name` field._

_See also [Kaniko](https://github.com/GoogleContainerTools/kaniko) and [BuildKit](https://github.com/moby/buildkit)._

## Cloud native

`Pipelines` are cloud native in that they:

* Run on kubernetes
* Have kubernetes clusters as a first class type (or will, see [#68](https://github.com/knative/build-pipeline/issues/68))
* Use containers as their building block