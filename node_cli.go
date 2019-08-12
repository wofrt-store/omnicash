/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
	cliconfig "slaveNode/config"
	"slaveNode/event"
	"github.com/spf13/cobra"
)

func newFabricCLICmd() *cobra.Command {
	logging.InitLogger(deflogger.LoggerProvider())

	mainCmd := &cobra.Command{
		Use: "node-cli",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	flags := mainCmd.PersistentFlags()
	cliconfig.InitConfigFile(flags)
	cliconfig.InitLoggingLevel(flags)
	cliconfig.InitUserName(flags)
	cliconfig.InitUserPassword(flags)
	cliconfig.InitOrdererTLSCertificate(flags)
	cliconfig.InitPrintFormat(flags)
	cliconfig.InitWriter(flags)
	cliconfig.InitBase64(flags)
	//cliconfig.InitPeerURL(flags,"grpcs://peer0.org2.daima.test:7051")
	cliconfig.InitPeerURL(flags,"grpcs://localhost:7051")
	cliconfig.InitOrgIDs(flags,"org2")
	cliconfig.InitChannelID(flags,"omni-main")
	mainCmd.AddCommand(event.Cmd())
	return mainCmd
}

func main() {
	if newFabricCLICmd().Execute() != nil {
		os.Exit(1)
	}
}
