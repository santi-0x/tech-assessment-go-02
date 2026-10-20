package access

import "go-relay/model"

type Roles struct {
	operator string
}

func New(operator string) (Roles, error) {
	if model.Blank(operator) {
		return Roles{}, model.ErrOperator
	}
	return Roles{operator: operator}, nil
}

func (r Roles) Require(caller string) error {
	if caller != r.operator {
		return model.ErrOperator
	}
	return nil
}
