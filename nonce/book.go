package nonce

import (
	"sync"

	"go-relay/model"
)

type Book struct {
	mu   sync.Mutex
	next map[string]uint64
}

func New() *Book {
	return &Book{next: map[string]uint64{}}
}

func (b *Book) Allocate(sender string) (uint64, error) {
	if model.Blank(sender) {
		return 0, model.ErrAddress
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	value := b.next[sender]
	b.next[sender] = value + 1
	return value, nil
}

func (b *Book) Commit(sender string, nonce uint64) (uint64, error) {
	if model.Blank(sender) {
		return 0, model.ErrAddress
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.next[sender] != nonce {
		return 0, model.ErrNonce
	}
	b.next[sender] = nonce + 1
	return nonce, nil
}
