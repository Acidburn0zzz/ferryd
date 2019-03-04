//
// Copyright © 2017-2019 Solus Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package jobs

import (
	"ferryd/core"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// DeltaPackageJobHandler is responsible for delta'ing a specific package in a specific repository
type DeltaPackageJobHandler struct {
	repoID string
	packageName string
}

// NewDeltaPackageJob will return a job suitable for adding to the job processor
func NewDeltaPackageJob(id string, packageName string) *JobEntry {
	return &JobEntry{
		sequential: true,
		Type:       DeltaPackage,
		Params:     []string{id, packageName,},
	}
}

// DeltaPackageJobHandlerJobHandler will create a job handler for the input job and ensure it validates
func NewDeltaPackageJobHandler(j *JobEntry) (*DeltaPackageJobHandler, error) {
	if len(j.Params) < 1 {
		return nil, fmt.Errorf("job has invalid parameters")
	}

	return &DeltaPackageJobHandler{
		repoID: j.Params[0],
		packageName: j.Params[1],
	}, nil
}

// Execute will delta a specific package in the given repository if possible
// Note that it will NOT index the repository, this is a separate step as it takes a significant amount of time to fully produce all initial
// deltas.
func (j *DeltaPackageJobHandler) Execute(jproc *Processor, manager *core.Manager) error {
	if packageMetas, err := manager.GetPackages(j.repoID, j.packageName); err == nil {
		if len(packageMetas) < 1 { // Doesn't exist
			log.WithFields(log.Fields{
				"repo": j.repoID,
				"package": j.packageName,
			}).Warning("Requested delta for package which does not exist")

			return nil
		}

		for _, meta := range packageMetas { // For each libeopkg.MetaPackage returned by GetPackages
			jproc.PushJob(NewDeltaPackageJob(j.repoID, meta.Name))
		}

		return nil
	} else {
		return err
	}
}

// Describe returns a human readable description for this job
func (j *DeltaPackageJobHandler) Describe() string {
	return fmt.Sprintf("Produce deltas for '%s' in '%s'", j.packageName, j.repoID)
}
