/*
Copyright © 2022 TWR Engineering

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

package resources

type PDBOutput struct {
	PDBs []PDB `json:"pdbs,omitempty"`
}

type PDB struct {
	Name               string   `json:"name"`
	Namespace          string   `json:"namespace"`
	Pods               []string `json:"pods"`
	Selectors          string   `json:"selectors"`
	DisruptionsAllowed int      `json:"disruptionsAllowed"`
	NewMinAvailable    string   `json:"newMinAvailable"`
	NewMaxUnavailable  string   `json:"newMaxUnavailable"`
	OldMinAvailable    string   `json:"oldMinAvailable"`
	OldMaxUnavailable  string   `json:"oldMaxUnavailable"`
}
