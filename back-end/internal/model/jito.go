package model

import (
	"github.com/gagliardetto/solana-go"
)

type JitoSchedule struct {
	NodePublicKey solana.PublicKey `json:"node_public_key"`
	RunningJito   bool             `json:"running_jito"`
}
