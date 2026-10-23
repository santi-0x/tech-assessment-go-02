package merkle

import "crypto/sha256"

func Pair(left, right [32]byte) [32]byte {
	sum := sha256.New()
	sum.Write([]byte{0x01})
	sum.Write(left[:])
	sum.Write(right[:])
	var out [32]byte
	copy(out[:], sum.Sum(nil))
	return out
}

func Verify(root, leaf [32]byte, siblings [][32]byte, index uint32) bool {
	if len(siblings) > 32 {
		return false
	}
	node := leaf
	for _, sibling := range siblings {
		if index&1 == 1 {
			node = Pair(sibling, node)
		} else {
			node = Pair(node, sibling)
		}
		index >>= 1
	}
	return index == 0 && node == root
}
