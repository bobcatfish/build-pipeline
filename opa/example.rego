package authz

pipelinerun[{
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
                            "name": myresource,
                        }
                    }
                ],
            }
        ],
    }
}] {
    data.kubernetes.pipelineresources["somens"][myresource]
}