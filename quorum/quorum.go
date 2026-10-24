package quorum

import (
	"sync"

	"go-relay/mac"
	"go-relay/model"
)

type Set struct {
	mu        sync.Mutex
	keys      map[string][32]byte
	threshold int
}

func New(threshold int) (*Set, error) {
	if threshold < 1 {
		return nil, model.ErrQuorum
	}
	return &Set{
		keys:      map[string][32]byte{},
		threshold: threshold,
	}, nil
}

func (s *Set) Register(name string, key [32]byte) error {
	if model.Blank(name) {
		return model.ErrAddress
	}
	var zero [32]byte
	if key == zero {
		return model.ErrQuorum
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.keys[name]; exists {
		return model.ErrQuorum
	}
	s.keys[name] = key
	return nil
}

func (s *Set) Accept(body []byte, votes []model.Vote) error {
	if len(body) == 0 {
		return model.ErrQuorum
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	good := 0
	for _, vote := range votes {
		key, ok := s.keys[vote.Signer]
		if !ok {
			continue
		}
		if !mac.Verify(key[:], body, vote.Tag[:]) {
			continue
		}
		good++
	}
	if good < s.threshold {
		return model.ErrQuorum
	}
	return nil
}
