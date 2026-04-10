package model

import "github.com/uptrace/bun"

type Exchange struct {
	bun.BaseModel `bun:"table:exchanges,alias:e" swaggerignore:"true"`

	ID   uint64 `json:"id" db:"id" bun:"id,pk,autoincrement" example:"1234567890"`
	Name string `json:"name" db:"name" bun:"name" example:"Binance"`
}

func (Exchange) TableName() string {
	return "exchanges"
}
