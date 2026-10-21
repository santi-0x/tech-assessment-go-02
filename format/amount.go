package format

import (
	"strconv"
	"strings"
)

func Group(asset string, amount uint64) string {
	raw := strconv.FormatUint(amount, 10)
	var b strings.Builder
	for i := 0; i < len(raw); i++ {
		if i > 0 && (len(raw)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(raw[i])
	}
	b.WriteByte(' ')
	b.WriteString(asset)
	return b.String()
}
