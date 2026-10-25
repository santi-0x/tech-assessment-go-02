package relay

import (
	"sync"

	"go-relay/access"
	"go-relay/broadcast"
	"go-relay/config"
	"go-relay/fee"
	"go-relay/header"
	"go-relay/inbox"
	"go-relay/merkle"
	"go-relay/model"
	"go-relay/nonce"
	"go-relay/payload"
	"go-relay/quorum"
	"go-relay/vault"
)

type Relay struct {
	cfg    *config.Config
	roles  access.Roles
	nonces *nonce.Book
	cursor *header.Cursor
	box    *inbox.Store
	guards *quorum.Set
	escrow *vault.Ledger
	log    *broadcast.Log
	mu     sync.Mutex
	paused bool
	feeBps uint32
	feeTo  string
	spent  map[[32]byte]struct{}
}

func New(cfg *config.Config) (*Relay, error) {
	if cfg == nil {
		return nil, model.ErrAddress
	}
	roles, err := access.New(cfg.Operator())
	if err != nil {
		return nil, err
	}
	guards, err := quorum.New(cfg.Threshold())
	if err != nil {
		return nil, err
	}
	return &Relay{
		cfg:    cfg,
		roles:  roles,
		nonces: nonce.New(),
		cursor: header.New(),
		box:    inbox.New(),
		guards: guards,
		escrow: vault.New(),
		log:    broadcast.New(),
		feeBps: cfg.FeeBps(),
		feeTo:  cfg.FeeRecipient(),
		spent:  map[[32]byte]struct{}{},
	}, nil
}

func (r *Relay) RegisterGuardian(caller, name string, key [32]byte) error {
	if err := r.roles.Require(caller); err != nil {
		return err
	}
	return r.guards.Register(name, key)
}

func (r *Relay) Advance(caller string, h model.Header) error {
	if err := r.roles.Require(caller); err != nil {
		return err
	}
	if h.Chain != r.cfg.SourceChain() {
		return model.ErrChain
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cursor.Advance(h)
}

func (r *Relay) Fund(caller string, amount uint64) error {
	if err := r.roles.Require(caller); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.escrow.Fund(amount)
}

func (r *Relay) SetPaused(caller string, paused bool) error {
	if err := r.roles.Require(caller); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paused = paused
	return nil
}

func (r *Relay) SetFeeBps(caller string, bps uint32) error {
	if err := r.roles.Require(caller); err != nil {
		return err
	}
	if bps > fee.MaxBps {
		return model.ErrFee
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.feeBps = bps
	return nil
}

func (r *Relay) SetFeeRecipient(caller, destination string) error {
	if err := r.roles.Require(caller); err != nil {
		return err
	}
	if model.Blank(destination) {
		return model.ErrAddress
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.feeTo = destination
	return nil
}

func (r *Relay) CommitBatch(caller string, root [32]byte, votes []model.Vote) error {
	if err := r.roles.Require(caller); err != nil {
		return err
	}
	var zero [32]byte
	if root == zero {
		return model.ErrRoot
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.paused {
		return model.ErrPaused
	}
	tip, ok := r.cursor.Tip(r.cfg.SourceChain())
	if !ok || tip.Number == 0 {
		return model.ErrHeader
	}
	body := payload.EncodeRoot(r.cfg.SourceChain(), r.cfg.DestChain(), root)
	if err := r.guards.Accept(body, votes); err != nil {
		return err
	}
	return r.box.Commit(root, tip.Number)
}

func (r *Relay) Prove(leaf [32]byte, siblings [][32]byte, index uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.paused {
		return model.ErrPaused
	}
	var zero [32]byte
	for _, root := range r.box.Roots() {
		if root == zero {
			continue
		}
		if merkle.Verify(root, leaf, siblings, index) {
			return r.box.Note(leaf, root)
		}
	}
	return model.ErrProof
}

func (r *Relay) Release(msg model.Message) (model.Receipt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.paused {
		return model.Receipt{}, model.ErrPaused
	}
	if err := validate(r.cfg, msg); err != nil {
		return model.Receipt{}, err
	}
	tip, ok := r.cursor.Tip(msg.SourceChain)
	if !ok {
		return model.Receipt{}, model.ErrHeader
	}
	if !header.Confirmed(tip.Number, msg.Block, r.cfg.Confirmations()) {
		return model.Receipt{}, model.ErrFinality
	}
	cut, err := fee.On(msg.Amount, r.feeBps)
	if err != nil {
		return model.Receipt{}, err
	}
	if cut > msg.Amount {
		return model.Receipt{}, model.ErrFee
	}
	net := msg.Amount - cut
	leaf := payload.Leaf(msg)
	if !r.box.Ready(leaf) {
		return model.Receipt{}, model.ErrProof
	}
	if _, seen := r.spent[leaf]; seen {
		return model.Receipt{}, model.ErrSpent
	}
	if r.escrow.Pool() < msg.Amount {
		return model.Receipt{}, model.ErrBalance
	}
	if _, err := r.nonces.Commit(msg.Sender, msg.Nonce); err != nil {
		return model.Receipt{}, err
	}
	if err := r.escrow.Settle(msg.Recipient, r.feeTo, msg.Amount, net, cut); err != nil {
		return model.Receipt{}, err
	}
	r.spent[leaf] = struct{}{}
	receipt := model.Receipt{
		Sender:       msg.Sender,
		Recipient:    msg.Recipient,
		Asset:        msg.Asset,
		Amount:       msg.Amount,
		Net:          net,
		Fee:          cut,
		FeeRecipient: r.feeTo,
		Nonce:        msg.Nonce,
		SourceChain:  msg.SourceChain,
		DestChain:    msg.DestChain,
		Block:        msg.Block,
		Leaf:         leaf,
	}
	r.log.Append(receipt)
	return receipt, nil
}

func (r *Relay) Balance(account string) uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.escrow.Balance(account)
}

func (r *Relay) Escrow() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.escrow.Pool()
}

func (r *Relay) Log() []model.Receipt {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.log.Items()
}

func validate(cfg *config.Config, msg model.Message) error {
	if model.Blank(msg.Sender) || model.Blank(msg.Recipient) || model.Blank(msg.Asset) {
		return model.ErrAddress
	}
	if msg.Amount == 0 || msg.Amount > cfg.MaxAmount() {
		return model.ErrAmount
	}
	if msg.SourceChain != cfg.SourceChain() || msg.DestChain != cfg.DestChain() {
		return model.ErrChain
	}
	return nil
}
