/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"
	"slaveNode/action"
	"slaveNode/config"
	cliconfig "slaveNode/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	ledgerUtil "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/ledger/util"
	fabricCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	fabriccmn "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
	"strings"
	"sync"
	"time"
)


type blockListenServer struct {
	action.Action
	inputEvent
}

func newBlockListenServer(flags *pflag.FlagSet) (*blockListenServer, error) {
	_action := &blockListenServer{inputEvent: inputEvent{done: make(chan bool)}}
	err := _action.Initialize(flags)
	return _action, err
}

var listenBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Listen to block events.",
	Long:  "Listen to block events",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newBlockListenServer(cmd.Flags())
		if err != nil {
			fmt.Println("Error while initializing listenBlockAction: %v", err)
			return
		}

		defer action.Terminate()
		err = action.invoke()
		if err != nil {
			fmt.Println("Error while running listenBlockAction: %v", err)
		}
	},
}

func getListenBlockCmd() *cobra.Command {
	flags := listenBlockCmd.Flags()
	cliconfig.InitPeerURL(flags)
	cliconfig.InitChannelID(flags)
	cliconfig.InitChaincodeID(flags)
	cliconfig.InitTimeout(flags)
	return listenBlockCmd
}


func (self *blockListenServer) invoke() error {
	channel := cliconfig.Config().ChannelID()
	eventHub, err := self.EventHub()
	if err != nil {
		return err
	}

	callback := func(block *fabricCommon.Block) {
		blockNum := block.Header.Number
		validMap := make(map[int]string)
		blockNumStr := fmt.Sprintf("%v", blockNum)
		numEnvelopes := len(block.Data.Data)
		if numEnvelopes == 0 {
			return
		}
		payload := utils.ExtractPayloadOrPanic(utils.ExtractEnvelopeOrPanic(block, 0))
		chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
		if err != nil {
			panic(err)
		}
		ccHdrExt := &pb.ChaincodeHeaderExtension{}
		unmarshalOrPanic(chdr.Extension, ccHdrExt)
		channel = chdr.ChannelId

		channelClient,err := self.GetOrgAdminChannelClient(chdr.ChannelId,self.OrgID()); if err!=nil{
			panic(errors.Errorf("获取客户端失败,channnel:",chdr.ChannelId," orgid:",self.OrgID()))
		}

		blockAfter, err := channelClient.QueryBlock(int(blockNum))
		if err != nil {
			fmt.Println("查询区块详情失败："+err.Error(), "块号：", blockNum)
		} else {
			block = blockAfter
		}

		txFilter := ledgerUtil.TxValidationFlags(block.Metadata.Metadata[fabriccmn.BlockMetadataIndex_TRANSACTIONS_FILTER])
		for i := 0; i < len(txFilter); i++ {
			validMap[i] = txFilter.Flag(i).String()
		}

		for i := 0; i < numEnvelopes; i++ {
			valid := strings.TrimSpace(strings.ToUpper(validMap[i]))
			payload := utils.ExtractPayloadOrPanic(utils.ExtractEnvelopeOrPanic(block, i))
			chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
			if err != nil {
				panic(err)
			}
			ccHdrExt := &pb.ChaincodeHeaderExtension{}
			unmarshalOrPanic(chdr.Extension, ccHdrExt)
			channel = chdr.ChannelId
			invokeTime := time.Unix(chdr.Timestamp.Seconds, 0).Format(config.YMRDHS_Format)

			if common.HeaderType(chdr.Type) == common.HeaderType_CONFIG {
				envelope := &common.ConfigEnvelope{}
				if err := proto.Unmarshal(payload.Data, envelope); err != nil {
					panic(errors.Errorf("Bad envelope: %v", err))
				}
			} else if common.HeaderType(chdr.Type) == common.HeaderType_ENDORSER_TRANSACTION {
				txId := chdr.TxId
				tx, err := utils.GetTransaction(payload.Data)
				if err != nil {
					fmt.Println("unmarshal transaction payload")
					return
				}
				wg := new(sync.WaitGroup)
				for _, action := range tx.Actions {
					chaPayload, err := utils.GetChaincodeActionPayload(action.Payload)
					if err != nil {
						panic(err)
					}
					prp := &pb.ProposalResponsePayload{}
					unmarshalOrPanic(chaPayload.Action.ProposalResponsePayload, prp)

					chaincodeAction := &pb.ChaincodeAction{}
					unmarshalOrPanic(prp.Extension, chaincodeAction)

					payloadStr := string(chaincodeAction.Response.Payload)
					if ok := strings.Contains(payloadStr, ";"); ok && len(strings.Split(payloadStr, ";")) > 2 {
						wg.Add(1)
						go HandlePayload(channel, payloadStr, txId, valid, invokeTime, blockNumStr, wg)
					}
				}
				//等待交易记录完成
				wg.Wait()
			}
		}
	}
	//事件注册
	eventHub.RegisterBlockEvent(callback)

	self.WaitForEnter()

	fmt.Printf("Unregistering block event\n")
	eventHub.UnregisterBlockEvent(callback)

	return nil
}

func unmarshalOrPanic(buf []byte, pb proto.Message) {
	err := proto.Unmarshal(buf, pb)
	if err != nil {
		panic(err)
	}
}

func getMetadataOrPanic(blockMetaData *fabriccmn.BlockMetadata, index fabriccmn.BlockMetadataIndex) *fabriccmn.Metadata {
	metaData := &fabriccmn.Metadata{}
	err := proto.Unmarshal(blockMetaData.Metadata[index], metaData)
	if err != nil {
		panic(errors.Errorf("Unable to unmarshal meta data at index %d", index))
	}
	return metaData
}


func HandlePayload(channel, payloadStr, txId, valid, invokeTime, blockNum string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("******************** 通道:",channel,", 块号:",blockNum,", payloadStr 数据为********************\n", payloadStr)
	resultArray := strings.Split(payloadStr, ";")
	resultStr := resultArray[2]
	result := strings.Split(resultStr, ",")
	//fmt.Println("数据长度,状态",len(result),valid)

	if len(result) > 11 && !gjson.Valid(resultStr)  && valid == "VALID"{
		//for i:=0;i<len(result);i++{
		//	fmt.Println(i," value:",result[i])
		//}
		chaincode := resultArray[0]
		ccfunc := resultArray[1]
		if chaincode == "token" && ccfunc == "transfer" {
			transAmount := result[9] 		//转出金额
			inAccount := result[8]			//转入账号
			outAccount := result[7]			//转出账号
			fee := result[4]				//手续费
			fmt.Println("交易ID:",txId)
			fmt.Println("交易时间:",invokeTime)
			fmt.Println("转出账号:",outAccount)
			fmt.Println("转入账号:",inAccount)
			fmt.Println("转出金额:",transAmount)
			fmt.Println("手续费:",fee)
			fmt.Println("区块号码:",blockNum)
		}
		return
	}
}