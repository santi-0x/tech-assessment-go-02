package broadcast

import "go-relay/model"

type Log struct {
	items []model.Receipt
}

func New() *Log {
	return &Log{}
}

func (l *Log) Append(receipt model.Receipt) {
	l.items = append(l.items, receipt)
}

func (l *Log) Items() []model.Receipt {
	out := make([]model.Receipt, len(l.items))
	copy(out, l.items)
	return out
}
