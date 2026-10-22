package inbox

import "go-relay/model"

var emptyRoot [32]byte

type Store struct {
	committed map[[32]byte]uint64
	proven    map[[32]byte][32]byte
}

func New() *Store {
	s := &Store{
		committed: map[[32]byte]uint64{},
		proven:    map[[32]byte][32]byte{},
	}
	s.committed[emptyRoot] = 1
	return s
}

func (s *Store) Commit(root [32]byte, at uint64) error {
	if root == emptyRoot || at == 0 {
		return model.ErrRoot
	}
	if _, exists := s.committed[root]; exists {
		return model.ErrRoot
	}
	s.committed[root] = at
	return nil
}

func (s *Store) Note(leaf, root [32]byte) error {
	if leaf == emptyRoot || root == emptyRoot {
		return model.ErrProof
	}
	at, ok := s.committed[root]
	if !ok || at == 0 {
		return model.ErrRoot
	}
	if _, exists := s.proven[leaf]; exists {
		return model.ErrSpent
	}
	s.proven[leaf] = root
	return nil
}

func (s *Store) Ready(leaf [32]byte) bool {
	root := s.proven[leaf]
	return s.committed[root] > 0
}

func (s *Store) Roots() [][32]byte {
	out := make([][32]byte, 0, len(s.committed))
	for root, at := range s.committed {
		if at == 0 {
			continue
		}
		out = append(out, root)
	}
	return out
}
