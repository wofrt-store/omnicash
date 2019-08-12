package main

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"fmt"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/spf13/pflag"
	"github.com/spf13/cobra/cobra/cmd"
)

func setupTestChannel() (*channel.Channel, error) {
	client := mocks.NewMockClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return channel.NewChannel("testChannel", client)
}

func setupTestClient() *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUser("test")
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return client
}


type queueAction struct {
	action.Action
	numInvoked uint32
	done       chan bool
}

func newInvokeQueueAction(flags *pflag.FlagSet) (*queueAction, error) {
	action := &queueAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func main() {
	//channel, error := setupTestChannel()
	//if error != nil {
	//	fmt.Println(error.Error())
	//}
	//badtxid1, _ := channel.ClientContext().NewTxnID()
	//fmt.Println(badtxid1)
	//fcClient := setupTestClient()
	//fmt.Println("2:")
	//fmt.Println(fcClient.NewTxnID())
	//
	 cliconfig.Config().UserName()

	action, _ := newInvokeQueueAction(nil)
	action.Client().NewTxnID()


}
