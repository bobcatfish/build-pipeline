# Demo yaml

This folder contains `yaml` CRDs which can be used to demonstrate the current state of the
art in `Pipeline` functionality.

Features we can demonstrate:

1. [Decoupling](#decoupling)
2. [Typing](#typing)
3. [Cloud native](#cloud-native)

## Setup

Before running any of these examples, you will need to
[setup the DEVELOPMENT requirements](https://github.com/knative/build-pipeline/blob/master/DEVELOPMENT.md#getting-started).

1. Create [two kubernetes clusters](https://github.com/knative/build-pipeline/blob/master/DEVELOPMENT.md#kubernetes-cluster)
   for yourself (one will be our `prod`, one will be `qa`).
2. Replace the values in [`pipelineparams-qa.yaml`](pipelineparams-qa.yaml) and
   [`pipelineparams-prod.yaml`] with:

   1. Your cluster endpoint
   2. The service account to use for that cluster
      ([must already exsit in the cluster](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/))

### Seeing logs

This is currently WIP, as of [#167](https://github.com/knative/build-pipeline/pull/167) we'll be
streaming the logs to a [`PersistentVolumeClaim`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/),
and to get these logs you would need to deploy another pod that reads from this volume.

## Decoupling

### Target environments

`Pipelines` are decoupled from the environments they are run against, for example you can take the
`Pipeline` `deploy-pipeline.yaml` and run it with two separate sets of `PipelineParams`, one of which
will deploy to one environment, and the other will deploy to another environment.

1. Run [`pipeline.yaml`](pipeline.yaml) against your "qa" env:

   ```bash
   pipelinerun-qa.yaml
   ```

2. Run [`pipeline.yaml`](pipeline.yaml) against your "prod" env:

   ```bash
   pipelinerun-prod.yaml
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
   ```

_Note that `PipelineRuns` must have unique names, so to re-run this you'll need to manually
change the `Name` field._

## Typing

One of the features of the Pipeline CRD is that we have types associated with `Resources`
that are common for CI/CD pipelines that use k8s, for example you can use different
`Task` definitions to produce `Images` without having to make significant changes.

1. Run [`pipeline.yaml`](pipeline.yaml) against your "qa" env:

   ```bash
   ```

2. Note that the `Pipeline` defined in [`pipeline2.yaml`](pipeline2.yaml) is the same
   as [`pipeline.yaml`](pipeline.yaml), however instead of referencing
   [the Task which builds with Kaniko](kaniko.yaml), it uses
   [the Task which builds with BuildKit](buildkit.yaml). Run it against your "qa" env:

   ```bash
   ```


_Note that `PipelineRuns` must have unique names, so to re-run this you'll need to manually
change the `Name` field._

_See also [Kaniko](https://github.com/GoogleContainerTools/kaniko) and [BuildKit](https://github.com/moby/buildkit)._

## Cloud native

`Pipelines` are cloud native in that they:

* Run on kubernetes
* Have kubernetes clusters as a first class type (or will, see [#68](https://github.com/knative/build-pipeline/issues/68))
* Use containers as their building block