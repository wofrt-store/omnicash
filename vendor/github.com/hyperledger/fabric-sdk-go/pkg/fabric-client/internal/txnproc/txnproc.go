/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package txnproc provides functionality for processing fabric transactions.
package txnproc

import (
	"sync"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"fmt"
	"time"
	"context"
	"strings"
)

var logger = logging.NewLogger("fabric_sdk_go")

// SendTransactionProposalToProcessors sends a TransactionProposal to ProposalProcessor.
// 向提案处理器 发送 事务提案。
func SendTransactionProposalToProcessors(proposal *apitxn.TransactionProposal, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, errors.New("signedProposal is required")
	}

	if len(targets) < 1 {
		return nil, errors.New("targets is required")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*apitxn.TransactionProposalResponse
	var wg sync.WaitGroup

	//循环所有peer分别发送事务提案
	for _, p := range targets {
		wg.Add(1)
		go func(processor apitxn.ProposalProcessor) {
			defer wg.Done()

			r, err := processor.ProcessTransactionProposal(*proposal)
			if err != nil {
				logger.Debugf("Received error response from txn proposal processing: %v", err)
				// Error is handled downstream.
			}

			tpr := apitxn.TransactionProposalResponse{
				TransactionProposalResult: r, Err: err}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, &tpr)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()
	return transactionProposalResponses, nil
}

// SendTransactionProposalToProcessors sends a TransactionProposal to ProposalProcessor.
func SendTransactionProposalToProcessorsNew(proposal *apitxn.TransactionProposal, targets []apitxn.ProposalProcessor, peers []string) ([]*apitxn.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, errors.New("signedProposal is required")
	}

	if len(targets) < 1 {
		return nil, errors.New("targets is required")
	}
	var responseMtx sync.Mutex
	var transactionProposalResponses []*apitxn.TransactionProposalResponse
	var wg sync.WaitGroup

	var peerList []string

	for _, p := range targets {
		wg.Add(1)
		go func(processor apitxn.ProposalProcessor) {
			//新增一个4秒超时的上下文 cliconfig.config.timeout 默认是5s
			ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
			go func() {
				for {
					select {
					case <-ctx.Done():
						defer func() {
							if err := recover(); err != nil && !strings.Contains(err.(string), "negative WaitGroup counter") {
								fmt.Println("*************recover 1 err*****************")
								fmt.Println(err)
							}
						}()
						//如果processor.ProcessTransactionProposal 阻塞，则将wg执行Done，则120行不再阻塞。继续执行后续代码
						wg.Done()
						return
					}
				}
			}()
			r, err := processor.ProcessTransactionProposal(*proposal)
			if err != nil {
				logger.Debugf("Received error response from txn proposal processing: %v", err)
			}
			//没问题的节点添加到peerList数组中
			peerList = append(peerList, r.Endorser)

			tpr := apitxn.TransactionProposalResponse{
				TransactionProposalResult: r, Err: err}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, &tpr)
			responseMtx.Unlock()
			defer func() {
				if err := recover(); err != nil && !strings.Contains(err.(string), "negative WaitGroup counter") {
					fmt.Println("*************recover 2 err*****************")
					fmt.Println(err)
				}
			}()
			wg.Done()
		}(p)
	}
	wg.Wait()

	return transactionProposalResponses, nil
}
