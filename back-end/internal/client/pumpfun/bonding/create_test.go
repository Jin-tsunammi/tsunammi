package bonding

import "testing"

func TestQuoteBuyExactSolInMinTokensOut(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		params         *SwapParams
		spendableSolIn uint64
		want           uint64
	}{
		{
			name: "no fees",
			params: &SwapParams{
				VirtualTokenReserves: 1_000,
				VirtualSolReserves:   100,
			},
			spendableSolIn: 101,
			want:           500,
		},
		{
			name: "protocol and creator fees",
			params: &SwapParams{
				VirtualTokenReserves: 1_000,
				VirtualSolReserves:   100,
				ProtocolFeeBps:       100,
				CreatorFeeBps:        100,
			},
			spendableSolIn: 100,
			want:           492,
		},
		{
			name: "ceil fee adjustment",
			params: &SwapParams{
				VirtualTokenReserves: 1_000,
				VirtualSolReserves:   100,
				ProtocolFeeBps:       1,
				CreatorFeeBps:        1,
			},
			spendableSolIn: 100,
			want:           492,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := quoteBuyExactSolInMinTokensOut(tt.params, tt.spendableSolIn)
			if err != nil {
				t.Fatalf("quoteBuyExactSolInMinTokensOut() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("quoteBuyExactSolInMinTokensOut() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestQuoteBuyExactSolInMinTokensOutRejectsZeroQuote(t *testing.T) {
	t.Parallel()

	_, err := quoteBuyExactSolInMinTokensOut(&SwapParams{
		VirtualTokenReserves: 1_000,
		VirtualSolReserves:   100,
	}, 1)
	if err == nil {
		t.Fatal("quoteBuyExactSolInMinTokensOut() error = nil, want error")
	}
}
