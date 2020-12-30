//
// Copyright © 2017-2020 Solus Project
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

package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/cmd"
	"github.com/getsolus/ferryd/api"
	"os"
)

// CherryPick fulfills the "cherry-pick" sub-command
var CherryPick = &cmd.CMD{
	Name:  "cherry-pick",
	Alias: "cp",
	Short: "Sync a single package from one repo to another",
	Args:  &CherryPickArgs{},
	Run:   CherryPickRun,
}

// CherryPickArgs are the arguments to the "cherry-pick" sub-command
type CherryPickArgs struct {
	Source  string `desc:"Repo to cherry-pick from"`
	Dest    string `desc:"Repo to cherry-pick into"`
	Package string `desc:"Package to cherry-pick"`
}

// CherryPickRun executes the "cherry-pick" sub-command
func CherryPickRun(r *cmd.RootCMD, c *cmd.CMD) {
	// Convert our flags
	flags := r.Flags.(*GlobalFlags)
	args := c.Args.(*CherryPickArgs)
	// Create a Client
	client := v1.NewClient(flags.Socket)
	defer client.Close()
	// Run the job
	d, j, err := client.CherryPick(args.Source, args.Dest, args.Package)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while cherry-picking: %v\n", err)
		os.Exit(1)
	}
	// Print the job summary
	j.Print()
	// Print the diff
	d.Print(os.Stdout, false, !flags.NoColor)
}
