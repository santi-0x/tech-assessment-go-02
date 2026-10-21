package fee

import (
	"math/bits"

	"go-relay/model"
)

const Basis = 10_000
const MaxBps = 100

func On(amount uint64, bps uint32) (uint64, error) {
	if bps > MaxBps {
		return 0, model.ErrFee
	}
	hi, lo := bits.Mul64(amount, uint64(bps))
	if hi != 0 {
		return 0, model.ErrFee
	}
	return lo / Basis, nil
}
