package types

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

// Account ...
type Account struct {
	ID       uuid.UUID                  `json:"id"`
	Balances map[Symbol]decimal.Decimal `json:"-"`
}

// NewAccount ...
func NewAccount() Account {
	return Account{
		ID: uuid.NewV4()}
}

type BalanceItem struct {
	ID        uuid.UUID       `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	Amount    decimal.Decimal `json:"amount"`
}

type AccountBalanceConfig struct {
	Account *Account
	Symbol  Symbol
}

type AccountRepo interface {
	Find(uuid.UUID) (*Account, error)
	Save(*Account) error
	Balances(*Account, Symbol) BalanceRepo
}

type BalanceRepo interface {
	GetBalance() (decimal.Decimal, error)
	UpdateBalance(decimal.Decimal) error
	FindHolds() ([]*BalanceItem, error)
	CreateHold(*BalanceItem) error
	DeleteHold(*BalanceItem) error
	FindPosts() ([]*BalanceItem, error)
	CreatePost(*BalanceItem) error
	DeletePost(*BalanceItem) error
}
