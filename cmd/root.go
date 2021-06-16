/*
Copyright © 2021 SUSE LLC

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

package cmd

import (
	"github.com/itxaka/luet-mtree/pkg/log"
	"github.com/spf13/cobra"
	"os"
)

// NewRootCmd represents the base command when called without any subcommands
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "luet-mtree [event] [payload]",
		Args:          cobra.ExactArgs(2),
		Short:         "Without a subcommand, luet mtree will parse events and their payloads from luet",
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := newEventCmd(args)
			if err != nil {
				log.Log(err.Error())
				os.Exit(1)
			}
		},
	}

	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newVersionCmd())

	return cmd
}
