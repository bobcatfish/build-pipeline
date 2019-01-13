package authz

# If there are no git resources, return nothing
test_create_nothing_for_no_resources {
    pipelinerun = set() with data.kubernetes.pipelineresources as {}
}

# If I have a git resource for which a pipelinerun should be created, give me a pipelinerun
test_create_pipelinerun_from_resource {
    x = { {
        "apiVersion": "pipeline.knative.dev/v1alpha1",
        "kind": "PipelineRun",
        "metadata": {
            "name": "demo-pipeline-run-1"
        },
        "spec": {
            "pipelineRef": {
            "name": "demo-pipeline"
            },
            "serviceAccount": "default",
            "trigger": {
                "type": "manual"
            },
            "resources": [
                {
                    "name": "build-skaffold-web",
                    "inputs": [
                        {
                            "name": "workspace",
                            "resourceRef": {
                                "name": "skaffold-git"
                            }
                        }
                    ],
                }
            ],
        }
    } }
    pipelinerun = x with data.kubernetes.pipelineresources as {
        "somens": {
            "skaffold-git": {
                "spec": {
                    "type": "git",
                    "params": [
                        {
                            "name": "revision",
                            "value": "master",
                        },
                        {
                            "name": "url",
                            "value": "https://github.com/GoogleContainerTools/skaffold",
                        },
                    ],
                }
            }
    }}
}