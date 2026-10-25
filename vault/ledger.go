package vault

import "go-relay/model"

type Ledger struct {
	pool     uint64
	accounts map[string]uint64
}

func New() *Ledger {
	return &Ledger{accounts: map[string]uint64{}}
}

func (l *Ledger) Fund(amount uint64) error {
	if amount == 0 {
		return model.ErrAmount
	}
	sum, ok := add(l.pool, amount)
	if !ok {
		return model.ErrAmount
	}
	l.pool = sum
	return nil
}

func (l *Ledger) Pool() uint64 {
	return l.pool
}

func (l *Ledger) Balance(account string) uint64 {
	return l.accounts[account]
}

func (l *Ledger) Settle(recipient, treasury string, gross, net, cut uint64) error {
	if model.Blank(recipient) || model.Blank(treasury) {
		return model.ErrAddress
	}
	sum, ok := add(net, cut)
	if !ok || sum != gross {
		return model.ErrAmount
	}
	if l.pool < gross {
		return model.ErrBalance
	}
	nextRecipient, ok := add(l.accounts[recipient], net)
	if !ok {
		return model.ErrAmount
	}
	nextTreasury := l.accounts[treasury]
	creditCut := cut
	if recipient == treasury {
		nextRecipient, ok = add(nextRecipient, cut)
		if !ok {
			return model.ErrAmount
		}
		creditCut = 0
	} else if cut > 0 {
		nextTreasury, ok = add(nextTreasury, cut)
		if !ok {
			return model.ErrAmount
		}
	}
	l.pool -= gross
	l.accounts[recipient] = nextRecipient
	if creditCut > 0 {
		l.accounts[treasury] = nextTreasury
	}
	return nil
}

func add(left, right uint64) (uint64, bool) {
	sum := left + right
	if sum < left {
		return 0, false
	}
	return sum, true
}
