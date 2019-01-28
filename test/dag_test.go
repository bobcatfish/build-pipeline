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
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	knativetest "github.com/knative/pkg/test"

	"github.com/knative/build-pipeline/pkg/apis/pipeline/v1alpha1"
)

const (
	// :((((((
	dagTimeout = time.Minute * 10
)

type fileTime struct {
	name string
	t    time.Time
}

type fileTimes []fileTime

func (f fileTimes) Len() int {
	return len(f)
}

func (f fileTimes) Less(i, j int) bool {
	return f[i].t.Before(f[j].t)
}

func (f fileTimes) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func getTimes(output string) (fileTimes, error) {
	times := fileTimes{}
	// The output of tail doesn't include the filename when it only outputs one file
	if len(output) <= 1 {
		return times, fmt.Errorf("output %q is too short to parse, this implies not all tasks wrote their files", output)
	}
	// ==> pipeline-task-1 <==
	//1544055212
	//
	//==> pipeline-task-2-parallel-1 <==
	//1544055304
	//
	//==> pipeline-task-2-parallel-2 <==
	//1544055263
	//
	//==> pipeline-task-3 <==
	//1544055368
	r, err := regexp.Compile("==> (.*) <==")
	if err != nil {
		return times, fmt.Errorf("couldn't compile filename regex: %v", err)
	}

	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i += 3 {
		// First line is the name of the file
		m := r.FindStringSubmatch(lines[i])
		if len(m) != 2 {
			return times, fmt.Errorf("didn't find expected filename in output line %d: %q", i, lines[i])
		}

		// Next line is the date
		i, err := strconv.Atoi(lines[i+1])
		if err != nil {
			return times, fmt.Errorf("error converting date %q to int: %v", lines[i+1], err)
		}

		times = append(times, fileTime{
			name: m[1],
			t:    time.Unix(int64(i), 0),
		})
	}
	return times, nil
}

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
	if _, err := c.PipelineClient.Create(DagPipeline(namespace)); err != nil {
		t.Fatalf("Failed to create pipeline Pipeline: %s", err)
	}
	if _, err := c.PipelineRunClient.Create(DagPipelineRun(namespace)); err != nil {
		t.Fatalf("Failed to create pipelineRun PipelineRun: %s", err)
	}

	logger.Infof("Waiting for DAG pipeline to complete")
	if err := WaitForPipelineRunState(c, "dag-pipeline-run", dagTimeout, func(tr *v1alpha1.PipelineRun) (bool, error) {
		c := tr.Status.GetCondition(duckv1alpha1.ConditionSucceeded)
		if c != nil {
			if c.IsTrue() {
				return true, nil
			} else if c.IsFalse() {
				return true, fmt.Errorf("Pipeline run failed with status %v", c.Status)
			}
		}
		return false, nil
	}, "PipelineRunSuccess"); err != nil {
		t.Fatalf("Error waiting for PipelineRun to finish: %s", err)
	}

	logger.Infof("Getting logs from results validation task")
	// The volume created with the results will have the same name as the TaskRun
	validationTaskRunName := "dag-pipeline-run-pipeline-task-4-validate-results"
	output, err := getBuildOutputFromVolume(logger, c, namespace, validationTaskRunName, "dag-validation-pod")
	if err != nil {
		t.Fatalf("Unable to get build output for taskrun %s: %s", validationTaskRunName, err)
	}
	fmt.Println(output)

	// Check that the overall order is correct
	times, err := getTimes(output)
	if err != nil {
		t.Fatalf("Unable to parse output %q: %v", output, err)
	}
	sort.Sort(times)

	if times[0].name != "pipeline-task-1" {
		t.Errorf("Expected first task to execute first, but %q was first", times[0].name)
	}
	if !strings.HasPrefix(times[1].name, "pipeline-task-2") {
		t.Errorf("Expected parallel tasks to run second & third, but %q was second", times[1].name)
	}
	if !strings.HasPrefix(times[2].name, "pipeline-task-2") {
		t.Errorf("Expected parallel tasks to run second & third, but %q was third", times[2].name)
	}
	if times[3].name != "pipeline-task-3" {
		t.Errorf("Expected third task to execute third, but %q was third", times[3].name)
	}

	// Check that the two tasks that can run in parallel did
	parallelDiff := times[2].t.Sub(times[1].t)
	if parallelDiff > (time.Second * 5) {
		t.Errorf("Expected parallel tasks to execute more or less at the ame time, but they were %v apart", parallelDiff)
	}
}

func DagPipeline(namespace string) *v1alpha1.Pipeline {
	return &v1alpha1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "dag-pipeline",
		},
		Spec: v1alpha1.PipelineSpec{
			// We intentionally mix up the order here so we can verify that the graph is being
			// created irrespective of the order the tasks are declared in
			Tasks: []v1alpha1.PipelineTask{{
				Name: "pipeline-task-3",
				TaskRef: v1alpha1.TaskRef{
					Name: "time-echo-task",
				},
				ResourceDependencies: []v1alpha1.ResourceDependency{{
					Name:       "folder",
					ProvidedBy: []string{"pipeline-task-2-parallel-1", "pipeline-task-2-parallel-2"},
				}},
				Params: []v1alpha1.Param{{
					Name:  "filename",
					Value: "pipeline-task-3",
				}, {
					Name:  "sleep-sec",
					Value: "5",
				}},
			}, {
				Name: "pipeline-task-2-parallel-2",
				TaskRef: v1alpha1.TaskRef{
					Name: "time-echo-task",
				},
				ResourceDependencies: []v1alpha1.ResourceDependency{{
					Name:       "folder",
					ProvidedBy: []string{"pipeline-task-1"},
				}},
				Params: []v1alpha1.Param{{
					Name:  "filename",
					Value: "pipeline-task-2-parallel-2",
				}, {
					Name:  "sleep-sec",
					Value: "5",
				}},
			}, {
				Name: "pipeline-task-2-parallel-1",
				TaskRef: v1alpha1.TaskRef{
					Name: "time-echo-task",
				},
				ResourceDependencies: []v1alpha1.ResourceDependency{{
					Name:       "folder",
					ProvidedBy: []string{"pipeline-task-1"},
				}},
				Params: []v1alpha1.Param{{
					Name:  "filename",
					Value: "pipeline-task-2-parallel-1",
				}, {
					Name:  "sleep-sec",
					Value: "5",
				}},
			}, {
				Name: "pipeline-task-1",
				TaskRef: v1alpha1.TaskRef{
					Name: "time-echo-task",
				},
				Params: []v1alpha1.Param{{
					Name:  "filename",
					Value: "pipeline-task-1",
				}, {
					Name:  "sleep-sec",
					Value: "5",
				}},
			}, {
				Name: "pipeline-task-4-validate-results",
				TaskRef: v1alpha1.TaskRef{
					Name: "folder-reader",
				},
				ResourceDependencies: []v1alpha1.ResourceDependency{{
					Name:       "folder",
					ProvidedBy: []string{"pipeline-task-3"},
				}},
				Params: []v1alpha1.Param{{
					Name:  "folder",
					Value: "output",
				}},
			}},
		},
	}
}

func DagPipelineRun(namespace string) *v1alpha1.PipelineRun {
	return &v1alpha1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "dag-pipeline-run",
		},
		Spec: v1alpha1.PipelineRunSpec{
			PipelineRef: v1alpha1.PipelineRef{
				Name: "dag-pipeline",
			},
			PipelineTriggerRef: v1alpha1.PipelineTriggerRef{
				Type: v1alpha1.PipelineTriggerTypeManual,
			},
			PipelineTaskResources: []v1alpha1.PipelineTaskResource{{
				Name: "pipeline-task-1",
				Inputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
				Outputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
			}, {
				Name: "pipeline-task-2-parallel-1",
				Inputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
				Outputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
			}, {
				Name: "pipeline-task-2-parallel-2",
				Inputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
				Outputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
			}, {
				Name: "pipeline-task-3",
				Inputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
				Outputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
			}, {
				Name: "pipeline-task-4-validate-results",
				Inputs: []v1alpha1.TaskResourceBinding{{
					Name: "folder",
					ResourceRef: v1alpha1.PipelineResourceRef{
						Name: "simple-repo",
					},
				}},
			}},
		},
	}
}

// TODO: could use volume instead?
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
					Name:        "folder",
					Description: "The folder to echo the results into",
					Default:     "output",
				}, {
					Name:        "sleep-sec",
					Description: "The number of seconds to sleep after echoing",
				}},
			},
			Outputs: &v1alpha1.Outputs{
				Resources: []v1alpha1.TaskResource{{
					Name: "folder",
					Type: v1alpha1.PipelineResourceTypeGit,
				}},
			},
			Steps: []corev1.Container{{
				Name:    "echo-time-into-file",
				Image:   "busybox",
				Command: []string{"sh"},
				// TODO: multiple steps maybe? but that seems slow?
				Args: []string{"-c", "pwd && mkdir -p ${inputs.params.folder} && date +%s > ${inputs.params.folder}/${inputs.params.filename} && sleep ${inputs.params.sleep-sec}"},
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
				Params: []v1alpha1.TaskParam{{
					Name:        "folder",
					Description: "The folder to read from",
				}},
			},
			Steps: []corev1.Container{{
				Name:    "read-all",
				Image:   "busybox",
				Command: []string{"sh"},
				// Display contents of all files, prefaced by their filenames
				Args: []string{"-c", "cd ${inputs.params.folder} && tail -n +1 -- *"},
			}},
		},
	}
}
