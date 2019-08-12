/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package apitxn allows SDK users to plugin their own implementations of transaction processing.
package apitxn

import (
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)


// ProposalProcessor simulates transaction proposal, so that a client can submit the result for ordering.
// 模拟交易提案，这样客户端就可以提交结果进行打包。
type ProposalProcessor interface {
	ProcessTransactionProposal(proposal TransactionProposal) (TransactionProposalResult, error)
}

// ProposalSender provides the ability for a transaction proposal to be created and sent.
// 提供创建和发送事务提案的能力。
type ProposalSender interface {
	SendTransactionProposal(ChaincodeInvokeRequest) ([]*TransactionProposalResponse, TransactionID, error)
}

// TransactionID contains the ID of a Fabric Transaction Proposal
// 包含一个Fabric事务提案的ID
type TransactionID struct {
	ID    string
	Nonce []byte
}

// ChaincodeInvokeRequest contains the parameters for sending a transaction proposal.
// 包含发送交易提案的参数。
type ChaincodeInvokeRequest struct {
	Targets      []ProposalProcessor
	ChaincodeID  string
	TxnID        TransactionID // TODO: does it make sense to include the TxnID in the request?
	TransientMap map[string][]byte
	Fcn          string
	Args         [][]byte
	SuperMiner   string
	NormalMiner  string
	PoolMiner	 string
	Peers        []string
}

// TransactionProposal requests simulation of a proposed transaction from transaction processors.
// 请求模拟来自事务处理器的拟议交易。
type TransactionProposal struct {
	TxnID TransactionID

	SignedProposal *pb.SignedProposal
	Proposal       *pb.Proposal
}

// TransactionProposalResponse encapsulates both the result of transaction proposal processing and errors.
// 封装了事务提案处理和错误的结果。
type TransactionProposalResponse struct {
	TransactionProposalResult
	Err error // TODO: consider refactoring
}

// TransactionProposalResult respresents the result of transaction proposal processing.
// 对事务提案处理的结果进行了响应。
type TransactionProposalResult struct {
	Endorser string
	Status   int32

	Proposal         TransactionProposal
	ProposalResponse *pb.ProposalResponse
}

// TODO: TransactionProposalResponse and TransactionProposalResult may need better names.
