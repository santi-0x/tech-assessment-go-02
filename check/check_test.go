package check

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sync"
	"testing"

	"go-relay/config"
	"go-relay/fee"
	relayhex "go-relay/hex"
	"go-relay/mac"
	"go-relay/merkle"
	"go-relay/model"
	"go-relay/nonce"
	"go-relay/payload"
	"go-relay/relay"
)

func TestRelay(t *testing.T) {
	checkMac(t)
	checkHex(t)
	checkFee(t)
	checkMerkle(t)
	checkBook(t)
	checkRelay(t)
}

func checkMac(t *testing.T) {
	t.Helper()
	key, err := relayhex.Decode("0x0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b")
	if err != nil || len(key) != 20 {
		t.Fatal(err)
	}
	tag := mac.Tag(key, []byte("Hi There"))
	want, err := relayhex.Decode("b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tag[:], want) {
		t.Fatalf("mac %x", tag)
	}
	if !mac.Verify(key, []byte("Hi There"), tag[:]) {
		t.Fatal("verify")
	}
}

func checkHex(t *testing.T) {
	t.Helper()
	raw, err := relayhex.Decode("0Xaa")
	if err != nil || len(raw) != 1 || raw[0] != 0xaa {
		t.Fatalf("hex %v %v", raw, err)
	}
	if _, err := relayhex.Decode("abc"); !errors.Is(err, model.ErrAmount) {
		t.Fatalf("odd hex %v", err)
	}
}

func checkFee(t *testing.T) {
	t.Helper()
	got, err := fee.On(10_000, 25)
	if err != nil || got != 25 {
		t.Fatalf("fee %d %v", got, err)
	}
	if _, err := fee.On(10_000, 101); !errors.Is(err, model.ErrFee) {
		t.Fatalf("cap %v", err)
	}
	if _, err := fee.On(^uint64(0), 100); !errors.Is(err, model.ErrFee) {
		t.Fatalf("overflow %v", err)
	}
}

func checkMerkle(t *testing.T) {
	t.Helper()
	leaves := [4][32]byte{
		sha256.Sum256([]byte{1}),
		sha256.Sum256([]byte{2}),
		sha256.Sum256([]byte{3}),
		sha256.Sum256([]byte{4}),
	}
	left := merkle.Pair(leaves[0], leaves[1])
	right := merkle.Pair(leaves[2], leaves[3])
	root := merkle.Pair(left, right)
	siblings := [][32]byte{leaves[3], left}
	if !merkle.Verify(root, leaves[2], siblings, 2) {
		t.Fatal("path")
	}
	if merkle.Verify(root, leaves[2], siblings, 0) {
		t.Fatal("index")
	}
	if !merkle.Verify(leaves[0], leaves[0], nil, 0) {
		t.Fatal("single")
	}
}

func checkBook(t *testing.T) {
	t.Helper()
	book := nonce.New()
	const threads = 8
	const calls = 200
	for _, sender := range []string{"alice", "bob"} {
		start := make(chan struct{})
		var wg sync.WaitGroup
		got := make([]uint64, threads*calls)
		var mu sync.Mutex
		var failed error
		for g := 0; g < threads; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				<-start
				for i := 0; i < calls; i++ {
					n, err := book.Allocate(sender)
					if err != nil {
						mu.Lock()
						failed = err
						mu.Unlock()
						return
					}
					got[g*calls+i] = n
				}
			}(g)
		}
		close(start)
		wg.Wait()
		if failed != nil {
			t.Fatal(failed)
		}
		seen := make(map[uint64]struct{}, len(got))
		min := got[0]
		max := got[0]
		for _, n := range got {
			if _, ok := seen[n]; ok {
				t.Fatalf("duplicate %s %d", sender, n)
			}
			seen[n] = struct{}{}
			if n < min {
				min = n
			}
			if n > max {
				max = n
			}
		}
		if len(seen) != threads*calls || min != 0 || max != uint64(threads*calls-1) {
			t.Fatalf("%s range %d..%d len %d", sender, min, max, len(seen))
		}
	}
	fresh := nonce.New()
	if _, err := fresh.Commit("carol", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.Commit("carol", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.Commit("carol", 1); !errors.Is(err, model.ErrNonce) {
		t.Fatalf("replay %v", err)
	}
	if _, err := fresh.Commit("dave", 0); err != nil {
		t.Fatal(err)
	}
}

func checkRelay(t *testing.T) {
	t.Helper()
	cfg, err := config.New("operator", 1, 10, 12, 25, "treasury", 1_000_000, 2)
	if err != nil {
		t.Fatal(err)
	}
	rel, err := relay.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	alpha := fill(0x11)
	beta := fill(0x22)
	gamma := fill(0x33)
	if err := rel.RegisterGuardian("stranger", "alpha", alpha); !errors.Is(err, model.ErrOperator) {
		t.Fatalf("register %v", err)
	}
	if err := rel.RegisterGuardian("operator", "alpha", alpha); err != nil {
		t.Fatal(err)
	}
	if err := rel.RegisterGuardian("operator", "beta", beta); err != nil {
		t.Fatal(err)
	}
	if err := rel.RegisterGuardian("operator", "gamma", gamma); err != nil {
		t.Fatal(err)
	}
	if err := rel.Fund("stranger", 10_000); !errors.Is(err, model.ErrOperator) {
		t.Fatalf("fund %v", err)
	}
	if err := rel.Fund("operator", 10_000); err != nil {
		t.Fatal(err)
	}
	if err := rel.SetPaused("stranger", true); !errors.Is(err, model.ErrOperator) {
		t.Fatalf("pause %v", err)
	}
	if err := rel.SetFeeBps("stranger", 40); !errors.Is(err, model.ErrOperator) {
		t.Fatalf("bps %v", err)
	}
	if err := rel.SetFeeBps("operator", 25); err != nil {
		t.Fatal(err)
	}
	if err := rel.SetFeeRecipient("stranger", "mallory"); !errors.Is(err, model.ErrOperator) {
		t.Fatalf("recipient %v", err)
	}
	if err := rel.SetFeeRecipient("operator", "treasury"); err != nil {
		t.Fatal(err)
	}
	if err := rel.Advance("stranger", model.Header{Chain: 1, Number: 0, Hash: headerHash(0)}); !errors.Is(err, model.ErrOperator) {
		t.Fatalf("advance %v", err)
	}
	if err := rel.Advance("operator", model.Header{Chain: 10, Number: 0, Hash: headerHash(0)}); !errors.Is(err, model.ErrChain) {
		t.Fatalf("dest header %v", err)
	}
	advanceTo(t, rel, 1, 100)
	parent := headerHash(100)
	broken := model.Header{Chain: 1, Number: 101, Parent: headerHash(50), Hash: headerHash(101)}
	if broken.Parent == parent {
		t.Fatal("fixture")
	}
	if err := rel.Advance("operator", broken); !errors.Is(err, model.ErrHeader) {
		t.Fatalf("parent %v", err)
	}

	msg := model.Message{
		Sender:      "alice",
		Recipient:   "bob",
		Asset:       "usdc",
		Amount:      10_000,
		Nonce:       0,
		SourceChain: 1,
		DestChain:   10,
		Block:       80,
	}
	shallow := msg
	shallow.Block = 100
	if _, err := rel.Release(shallow); !errors.Is(err, model.ErrFinality) {
		t.Fatalf("head block %v", err)
	}
	foreign := msg
	foreign.DestChain = 11
	if _, err := rel.Release(foreign); !errors.Is(err, model.ErrChain) {
		t.Fatalf("chain %v", err)
	}
	if err := rel.SetPaused("operator", true); err != nil {
		t.Fatal(err)
	}
	if _, err := rel.Release(msg); !errors.Is(err, model.ErrPaused) {
		t.Fatalf("paused %v", err)
	}
	if err := rel.SetPaused("operator", false); err != nil {
		t.Fatal(err)
	}

	leaf := payload.Leaf(msg)
	siblingMsg := msg
	siblingMsg.Sender = "carol"
	sibling := payload.Leaf(siblingMsg)
	root := merkle.Pair(leaf, sibling)
	body := payload.EncodeRoot(1, 10, root)
	votes := []model.Vote{
		{Signer: "alpha", Tag: mac.Tag(alpha[:], body)},
		{Signer: "beta", Tag: mac.Tag(beta[:], body)},
	}
	if err := rel.CommitBatch("stranger", root, votes); !errors.Is(err, model.ErrOperator) {
		t.Fatalf("commit role %v", err)
	}
	if err := rel.CommitBatch("operator", root, votes[:1]); !errors.Is(err, model.ErrQuorum) {
		t.Fatalf("one vote %v", err)
	}
	bad := []model.Vote{votes[0], votes[1]}
	bad[1].Tag[0] ^= 0xff
	if err := rel.CommitBatch("operator", root, bad); !errors.Is(err, model.ErrQuorum) {
		t.Fatalf("bad tag %v", err)
	}
	unknown := []model.Vote{votes[0], {Signer: "mallory", Tag: votes[0].Tag}}
	if err := rel.CommitBatch("operator", root, unknown); !errors.Is(err, model.ErrQuorum) {
		t.Fatalf("unknown %v", err)
	}
	if err := rel.CommitBatch("operator", root, votes); err != nil {
		t.Fatal(err)
	}
	if err := rel.Prove(leaf, [][32]byte{sibling}, 0); err != nil {
		t.Fatal(err)
	}
	receipt, err := rel.Release(msg)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Net != 9_975 || receipt.Fee != 25 || receipt.FeeRecipient != "treasury" {
		t.Fatalf("receipt %+v", receipt)
	}
	if receipt.Leaf != leaf || receipt.Recipient != "bob" || receipt.Nonce != 0 {
		t.Fatalf("receipt body %+v", receipt)
	}
	if rel.Balance("bob") != 9_975 || rel.Balance("treasury") != 25 || rel.Escrow() != 0 {
		t.Fatalf("balances bob %d treasury %d escrow %d", rel.Balance("bob"), rel.Balance("treasury"), rel.Escrow())
	}
	if len(rel.Log()) != 1 {
		t.Fatalf("log %d", len(rel.Log()))
	}
	if _, err := rel.Release(msg); !errors.Is(err, model.ErrSpent) {
		t.Fatalf("replay %v", err)
	}
}

func advanceTo(t *testing.T, rel *relay.Relay, chain, target uint64) {
	t.Helper()
	var parent [32]byte
	for n := uint64(0); n <= target; n++ {
		hash := headerHash(n)
		err := rel.Advance("operator", model.Header{
			Chain:  chain,
			Number: n,
			Parent: parent,
			Hash:   hash,
		})
		if err != nil {
			t.Fatalf("header %d: %v", n, err)
		}
		parent = hash
	}
}

func headerHash(n uint64) [32]byte {
	var out [32]byte
	out[0] = 0xa1
	binary.BigEndian.PutUint64(out[24:], n+1)
	return out
}

func fill(b byte) [32]byte {
	var out [32]byte
	for i := range out {
		out[i] = b
	}
	return out
}
