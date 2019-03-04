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

var copySourceCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy packages by source name",
	Long:  "Remove an existing package set in the ferryd instance",
	Run:   copySource,
	Aliases: []string{ "cherrypick", },
}

var (
	copyFromRepo string
	copyToRepo string
	copyPackageName string
	copyRelNum int
)

func init() {
	copySourceCmd.Flags().StringVarP(&copyFromRepo, "from", "f", "unstable", "From / Source Repository")
	copySourceCmd.Flags().StringVarP(&copyToRepo, "to", "t", "shannon", "To / Destination Repository")
	copySourceCmd.Flags().StringVarP(&copyPackageName, "package", "p", "", "Package Name")
	copySourceCmd.Flags().IntVarP(&copyRelNum, "release", "r", -1, "Release Number")
	copySourceCmd.MarkFlagRequired("package")

	RootCmd.AddCommand(copySourceCmd)
}

func copySource(cmd *cobra.Command, args []string) {
	for _, val := range []string{ copyFromRepo, copyToRepo, copyPackageName, } {
		if val == "" {
			fmt.Println(cmd.UsageString())
			return
		}
	}

	client := libferry.NewClient(socketPath)
	defer client.Close()

	if err := client.CopySource(copyFromRepo, copyToRepo, copyPackageName, copyRelNum); err != nil {
		fmt.Fprintf(os.Stderr, "Error while copying source: %v\n", err)
		return
	}
}
