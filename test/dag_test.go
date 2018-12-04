// +build e2e

/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knativetest "github.com/knative/pkg/test"

	"github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
)

func TestDAG(t *testing.T) {
	logger := getContextLogger(t.Name())
	c, namespace := setup(t, logger)

	knativetest.CleanupOnInterrupt(func() { tearDown(t, logger, c, namespace) }, logger)
	defer tearDown(t, logger, c, namespace)

	if _, err := c.TaskClient.Create(TimeEchoTask(namespace)); err != nil {
		t.Fatalf("Failed to create time echo Task: %s", err)
	}
	if _, err := c.TaskClient.Create(FolderReaderTask(namespace)); err != nil {
		t.Fatalf("Failed to create folder reader Task: %s", err)
	}
	if _, err := c.PipelineResourceClient.Create(SimpleRepo(namespace)); err != nil {
		t.Fatalf("Failed to create simple repo PipelineResource: %s", err)
	}

	// Create a task Diamond in a Pipeline - intentionally do it in the wrong order!
	// Run it with the Resource

	// Verify that times are within x of when they should be (parallels are within paralell)
	// Get logs
	// Look at logs
}

func SimpleRepo(namespace string) *v1alpha1.PipelineResource {
	// Input -> output sharing is currently only supported for resource type Git; this test really just needs
	// any volume where it can put data
	return &v1alpha1.PipelineResource{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "simple-repo",
		},
		Spec: v1alpha1.PipelineResourceSpec{
			Type: v1alpha1.PipelineResourceTypeGit,
			Params: []v1alpha1.Param{{
				Name:  "Url",
				Value: "https://github.com/githubtraining/example-basic",
			}},
		},
	}
}

func TimeEchoTask(namespace string) *v1alpha1.Task {
	return &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "time-echo-task",
		},
		Spec: v1alpha1.TaskSpec{
			// TODO(#124): we only want to write to this, maybe it should just be an output?
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name: "folder",
					Type: v1alpha1.PipelineResourceTypeGit,
				}},
				Params: []v1alpha1.TaskParam{{
					Name:        "filename",
					Description: "The name of the File to echo the time into",
				}, {
					Name:        "sleep-sec",
					Description: "The number of seconds to sleep after echoing",
				}},
			},
			Steps: []corev1.Container{{
				Name:    "echo-time-into-file",
				Image:   "busybox",
				Command: []string{"bash"},
				Args:    []string{"-c", "date +%s > ${inputs.params.folder} && sleep ${inputs.params.sleep-sec}"},
			}},
		},
	}
}

func FolderReaderTask(namespace string) *v1alpha1.Task {
	return &v1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "folder-reader",
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{{
					Name: "folder",
					Type: v1alpha1.PipelineResourceTypeGit,
				}},
			},
			Steps: []corev1.Container{{
				Name:    "read-all",
				Image:   "busybox",
				Command: []string{"bash"},
				// Display contents of all files, prefaced by their filenames
				Args: []string{"-c", " tail -n +1 -- *"},
			}},
		},
	}
}
