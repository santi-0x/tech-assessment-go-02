package config

import (
	"go-relay/fee"
	"go-relay/model"
)

type Config struct {
	operator      string
	sourceChain   uint64
	destChain     uint64
	confirmations uint64
	feeBps        uint32
	feeRecipient  string
	maxAmount     uint64
	threshold     int
}

func New(
	operator string,
	sourceChain uint64,
	destChain uint64,
	confirmations uint64,
	feeBps uint32,
	feeRecipient string,
	maxAmount uint64,
	threshold int,
) (*Config, error) {
	if model.Blank(operator) || model.Blank(feeRecipient) {
		return nil, model.ErrAddress
	}
	if sourceChain == 0 || destChain == 0 || sourceChain == destChain {
		return nil, model.ErrChain
	}
	if confirmations == 0 {
		return nil, model.ErrFinality
	}
	if feeBps > fee.MaxBps {
		return nil, model.ErrFee
	}
	if maxAmount == 0 {
		return nil, model.ErrAmount
	}
	if threshold < 1 {
		return nil, model.ErrQuorum
	}
	return &Config{
		operator:      operator,
		sourceChain:   sourceChain,
		destChain:     destChain,
		confirmations: confirmations,
		feeBps:        feeBps,
		feeRecipient:  feeRecipient,
		maxAmount:     maxAmount,
		threshold:     threshold,
	}, nil
}

func (c *Config) Operator() string      { return c.operator }
func (c *Config) SourceChain() uint64   { return c.sourceChain }
func (c *Config) DestChain() uint64     { return c.destChain }
func (c *Config) Confirmations() uint64 { return c.confirmations }
func (c *Config) FeeBps() uint32        { return c.feeBps }
func (c *Config) FeeRecipient() string  { return c.feeRecipient }
func (c *Config) MaxAmount() uint64     { return c.maxAmount }
func (c *Config) Threshold() int        { return c.threshold }
