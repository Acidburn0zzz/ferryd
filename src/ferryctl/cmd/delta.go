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

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"libferry"
	"os"
)

var deltaCmd = &cobra.Command{
	Use:   "delta [repoName]",
	Short: "Create deltas",
	Long:  "Rebuild all deltas for a repo",
	Run:   delta,
}

var repoName string
var packageName string

func init() {
	deltaCmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository to delta")
	deltaCmd.Flags().StringVarP(&packageName, "package", "p", "", "Package to delta")
	deltaCmd.MarkFlagRequired("repo")

	RootCmd.AddCommand(deltaCmd)
}

func delta(cmd *cobra.Command, args []string) {
	if repoName == "" {
		return
	}

	client := libferry.NewClient(socketPath)
	defer client.Close()

	if packageName == "" { // Delta entire repository
		if err := client.DeltaRepo(repoName); err != nil {
			fmt.Fprintf(os.Stderr, "Error while creating deltas: %v\n", err)
		}
	} else { // Package set, delta only specific package
		if err := client.DeltaPackage(repoName, packageName); err != nil {
			fmt.Fprintf(os.Stderr, "Error while creating delta for package %s in %s: %v\n", packageName, repoName, err)
		}
	}

	return
}
