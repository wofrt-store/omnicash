package config

import (
	"math/big"
	"time"
)

const (
	YMRDHS_Format                = "2006-01-02 15:04:05"
)

type RwSet struct {
	Key       string
	Namespace string
	RBlockNum string
	RTxNum    string
	WIsDelete bool
	WValue    string
}

type TxBlock struct {
	TxId      string
	Timestamp string
	BNum      *big.Int
	RwSets    []RwSet
}

type DbStruct struct {
	txId      string
	timestamp string
	bNum      *big.Int
	key       string
	namespace string
	rBlockNum *big.Int
	rTxNum    *big.Int
	wIsDelete bool
	wValue    string
}

type Blockinfo struct {
	num        *big.Int
	txId       string
	namespace  string
	key        string
	rBlockNum  *big.Int
	rTxNum     string
	wValue     string
	wValueBs64 string
	wIsDelete  string
	time       time.Time
	uptime     time.Time
}
