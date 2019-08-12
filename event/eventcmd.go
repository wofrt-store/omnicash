/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	cliconfig "slaveNode/config"
	"github.com/spf13/cobra"
)

var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Event commands",
	Long:  "Event commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Cmd returns the events command
func Cmd() *cobra.Command {
	cliconfig.InitChannelID(eventCmd.Flags())
	eventCmd.AddCommand(getListenBlockCmd())

	return eventCmd
}
