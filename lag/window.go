package lag

type Window struct {
	samples []uint64
	limit   int
}

func New(limit int) *Window {
	if limit < 1 {
		limit = 1
	}
	return &Window{limit: limit}
}

func (w *Window) Observe(delay uint64) {
	if len(w.samples) == w.limit {
		w.samples = w.samples[1:]
	}
	w.samples = append(w.samples, delay)
}

func (w *Window) Mean() uint64 {
	if len(w.samples) == 0 {
		return 0
	}
	var total uint64
	for _, sample := range w.samples {
		total += sample
	}
	return total / uint64(len(w.samples))
}
