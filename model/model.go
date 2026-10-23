package model

import "strings"

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrPaused   Error = "paused"
	ErrAddress  Error = "address"
	ErrAmount   Error = "amount"
	ErrChain    Error = "chain"
	ErrFee      Error = "fee"
	ErrFinality Error = "finality"
	ErrQuorum   Error = "quorum"
	ErrProof    Error = "proof"
	ErrNonce    Error = "nonce"
	ErrOperator Error = "operator"
	ErrBalance  Error = "balance"
	ErrHeader   Error = "header"
	ErrRoot     Error = "root"
	ErrSpent    Error = "spent"
)

type Header struct {
	Chain  uint64
	Number uint64
	Parent [32]byte
	Hash   [32]byte
}

type Vote struct {
	Signer string
	Tag    [32]byte
}

type Message struct {
	Sender      string
	Recipient   string
	Asset       string
	Amount      uint64
	Nonce       uint64
	SourceChain uint64
	DestChain   uint64
	Block       uint64
}

type Receipt struct {
	Sender       string
	Recipient    string
	Asset        string
	Amount       uint64
	Net          uint64
	Fee          uint64
	FeeRecipient string
	Nonce        uint64
	SourceChain  uint64
	DestChain    uint64
	Block        uint64
	Leaf         [32]byte
}

func Blank(value string) bool {
	return strings.TrimSpace(value) == ""
}
