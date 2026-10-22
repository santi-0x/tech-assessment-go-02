package header

import "go-relay/model"

type Cursor struct {
	tips map[uint64]model.Header
}

func New() *Cursor {
	return &Cursor{tips: map[uint64]model.Header{}}
}

func (c *Cursor) Advance(h model.Header) error {
	if h.Chain == 0 {
		return model.ErrChain
	}
	var zero [32]byte
	if h.Hash == zero {
		return model.ErrHeader
	}
	prev, ok := c.tips[h.Chain]
	if !ok {
		if h.Number != 0 {
			return model.ErrHeader
		}
		c.tips[h.Chain] = h
		return nil
	}
	if h.Number != prev.Number+1 {
		return model.ErrHeader
	}
	if h.Parent != prev.Hash || h.Hash == h.Parent {
		return model.ErrHeader
	}
	c.tips[h.Chain] = h
	return nil
}

func (c *Cursor) Tip(chain uint64) (model.Header, bool) {
	h, ok := c.tips[chain]
	return h, ok
}

func Confirmed(head, block, confirmations uint64) bool {
	if confirmations == 0 || head < block {
		return false
	}
	return head-block >= 1
}
