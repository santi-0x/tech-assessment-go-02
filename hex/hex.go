package hex

import "go-relay/model"

func Decode(text string) ([]byte, error) {
	body := text
	if len(body) >= 2 && body[0] == '0' && (body[1] == 'x' || body[1] == 'X') {
		body = body[2:]
	}
	if body == "" || len(body)%2 == 1 {
		return nil, model.ErrAmount
	}
	out := make([]byte, len(body)/2)
	for i := 0; i < len(body); i += 2 {
		hi, ok := digit(body[i])
		if !ok {
			return nil, model.ErrAmount
		}
		lo, ok := digit(body[i+1])
		if !ok {
			return nil, model.ErrAmount
		}
		out[i/2] = hi<<4 | lo
	}
	return out, nil
}

func digit(ch byte) (byte, bool) {
	switch {
	case ch >= '0' && ch <= '9':
		return ch - '0', true
	case ch >= 'a' && ch <= 'f':
		return ch - 'a' + 10, true
	case ch >= 'A' && ch <= 'F':
		return ch - 'A' + 10, true
	default:
		return 0, false
	}
}
