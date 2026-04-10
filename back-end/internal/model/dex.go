package model

import "github.com/gagliardetto/solana-go"

type PoolResponse struct {
	ProgramID string   `json:"programId"`
	ID        string   `json:"id"`
	MintA     MintInfo `json:"mintA"`
	MintB     MintInfo `json:"mintB"`
	FeeRate   float64  `json:"feeRate"`
	TVL       float64  `json:"tvl"`

	// Type               string       `json:"type"`
	// Price              float64      `json:"price"`
	// MintAmountA        float64      `json:"mintAmountA"`
	// MintAmountB        float64      `json:"mintAmountB"`
	// OpenTime           string       `json:"openTime"`
	// Day                Period       `json:"day"`
	// Week               Period       `json:"week"`
	// Month              Period       `json:"month"`
	// PoolType           []string     `json:"pooltype"`
	// RewardDefaultInfos []RewardInfo `json:"rewardDefaultInfos"`
	// RewardDefaultPool  string       `json:"rewardDefaultPoolInfos"`
	// FarmUpcomingCount  int          `json:"farmUpcomingCount"`
	// FarmOngoingCount   int          `json:"farmOngoingCount"`
	// FarmFinishedCount  int          `json:"farmFinishedCount"`
	// MarketID           string       `json:"marketId"`
	// LPMint             MintInfo     `json:"lpMint"`
	// LPPrice            float64      `json:"lpPrice"`
	// LPAmount           float64      `json:"lpAmount"`
	// BurnPercent        float64      `json:"burnPercent"`
	// LaunchMigratePool  bool         `json:"launchMigratePool"`
}

func (p *PoolResponse) IsPoolTypeOneOf(programID ...solana.PublicKey) bool {
	if programID == nil {
		return true
	}

	for _, id := range programID {
		if p.ProgramID == id.String() {
			return true
		}
	}

	return false
}

func (p *PoolResponse) IsCorrectMintOrder(mintA, mintB solana.PublicKey) bool {
	return p.MintA.Address == mintA.String() && p.MintB.Address == mintB.String()
}

type MintInfo struct {
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`

	// ChainID    int            `json:"chainId"`
	// ProgramID  string         `json:"programId"`
	// LogoURI    string         `json:"logoURI"`
	// Symbol     string         `json:"symbol"`
	// Name       string         `json:"name"`
	// Tags       []string       `json:"tags"`
	// Extensions map[string]any `json:"extensions"`
}

type Period struct {
	Volume      float64   `json:"volume"`
	VolumeQuote float64   `json:"volumeQuote"`
	VolumeFee   float64   `json:"volumeFee"`
	Apr         float64   `json:"apr"`
	FeeApr      float64   `json:"feeApr"`
	PriceMin    float64   `json:"priceMin"`
	PriceMax    float64   `json:"priceMax"`
	RewardApr   []float64 `json:"rewardApr"`
}

type RewardInfo struct {
	Mint      MintInfo `json:"mint"`
	PerSecond string   `json:"perSecond"`
	StartTime string   `json:"startTime"`
	EndTime   string   `json:"endTime"`
}
