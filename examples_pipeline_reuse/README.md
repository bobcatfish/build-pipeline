# Pipeline Reuse examples

A user should be able to take an existing `Pipeline` and run it against their own setup without
modifying the `Pipeline` itself, meaning using their own:

1. GitHub fork / branch (e.g. for PR testing)
2. Image registry
3. k8s cluster (to deploy to)

In the current implementation, to change (3), a user would create their own `PipelineParams`, so they
would not need to change the `Pipeline`.

But to change (1) or (2), they would have to create new `Resources` with their specific configuration.
Since `Resources` are bound in the `Pipeline`, this means they would need to change the `Pipeline` as
well.

We are considering moving `clusters` out of `PipelineParams` and into a `Resource` type, but this
would mean that for a user to use their own cluster, they would again need to change the `Pipeline`. 

## Solutions?

Some possible solutions for this problem:

1. We expand `PipelineParams` to include a generic list of keys and values, which can be templated
   inside all types (e.g. `Pipelines`, `Resources`, etc.).

   See [example_option_1](./example_option_1):

   * `pipelines/kritis-resources.yaml` contains templating for values like URL and revision
     in the `Resources`
   * `pipelineparams.yaml` contains the values for these params

   _This is very similar to option 3, the difference is that in this version we still have
   `Resource` CRDs, which would need to be updated on any changes to `PipelineParams`.

2. `PipelineParams` can override values inside of `Resources`.

   _This would be basically the same as option 1, except limited to `Resources` instead of working
   for any CRD._

3. Instead of declaring `Resources` as separate CRDs, we define them all inside `PipelineParams`

   See [example_option_3](./example_option_3):

   * `pipelineparams.yaml` now contains all Resource definitions
   * Resources are no longer their own type of CRD

4. Instead of binding `Resources` in `Pipeline`, we bind them in `PipelineRun` (and `TaskRun`)

   See [example_option_4](./example_option_4):

   * `kritis-pipeline-run.yaml` now contains all resource bindings
   * `pipeline/kritis.yaml` still contains the `passedConstraints` so that it
     can express the order

   Cons:

   - Resources are now handled in both the Pipeline _and_ the PipelineRun

5. A combination of (4) + (2): if you _want_ to override a value, you can do that by binding an overridden resource in (4).

   _We ruled this one out because we feel it is too confusing._


6. No `PipelineParams`; people need to change the Pipeline. Move the ServiceAccount and Result store information into the Pipeline itself.

   See [example_option_6](./example_option_6):

   * The `PipelineParams` are gone, values are moved into `pipeline/kritis.yaml`
   * The `PipelineRun` in `invocations/kritis-pipeline-run.yaml` now has very
    little in it.

   Cons:

   - Still need to change 2 things: create a Resource with the configuration you
  want, and change the Pipeline to use that Resoruce

7. Either (1) or (2) but instead of using `PipelineParams`, use `PipelineRun`
   and `TaskRun`.