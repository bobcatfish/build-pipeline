/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either extress or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package list

func listExtra(left, right []string) []string {
	extra := []string{}
	for _, s := range left {
		found := false
		for _, s2 := range right {
			if s == s2 {
				found = true
			}
		}
		if !found {
			extra = append(extra, s)
		}
	}
	return extra
}

// Diff compares two lists of strings and returns any strings in the which
// are present in one list and not the other.
func Diff(left, right []string) ([]string, []string) {
	extraLeft := listExtra(left, right)
	extraRight := listExtra(right, left)
	return extraLeft, extraRight
}
