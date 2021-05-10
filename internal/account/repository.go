package account

import (
	"time"

	"github.com/easterthebunny/spew-order/pkg/types"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

func NewBalanceItem(amt decimal.Decimal) BalanceItem {
	return BalanceItem{
		ID:        uuid.NewV4(),
		Timestamp: time.Now(),
		Amount:    amt}
}

type BalanceItem struct {
	ID        uuid.UUID       `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Amount    decimal.Decimal `json:"amount"`
}

type AccountRepository interface {
	Find(uuid.UUID) (*types.Account, error)
	Save(*types.Account) error
	Balances(*types.Account, types.Symbol) BalanceRepository
}

type BalanceRepository interface {
	GetBalance() (decimal.Decimal, error)
	UpdateBalance(decimal.Decimal) error
	FindHolds() ([]*BalanceItem, error)
	CreateHold(*BalanceItem) error
	DeleteHold(*BalanceItem) error
	FindPosts() ([]*BalanceItem, error)
	CreatePost(*BalanceItem) error
	DeletePost(*BalanceItem) error
}
