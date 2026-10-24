package payload

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"go-relay/model"
)

func Encode(msg model.Message) []byte {
	var buf bytes.Buffer
	buf.WriteString("go-relay/v1/lock")
	putU64(&buf, msg.SourceChain)
	putU64(&buf, msg.DestChain)
	putStr(&buf, msg.Sender)
	putStr(&buf, msg.Recipient)
	putStr(&buf, msg.Asset)
	putU64(&buf, msg.Amount)
	putU64(&buf, msg.Nonce)
	putU64(&buf, msg.Block)
	return buf.Bytes()
}

func EncodeRoot(source, dest uint64, root [32]byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("go-relay/v1/batch")
	putU64(&buf, source)
	putU64(&buf, dest)
	buf.Write(root[:])
	return buf.Bytes()
}

func Leaf(msg model.Message) [32]byte {
	sum := sha256.New()
	sum.Write([]byte{0x00})
	sum.Write(Encode(msg))
	var out [32]byte
	copy(out[:], sum.Sum(nil))
	return out
}

func putStr(buf *bytes.Buffer, value string) {
	raw := []byte(value)
	var n [4]byte
	binary.BigEndian.PutUint32(n[:], uint32(len(raw)))
	buf.Write(n[:])
	buf.Write(raw)
}

func putU64(buf *bytes.Buffer, value uint64) {
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], value)
	buf.Write(n[:])
}
