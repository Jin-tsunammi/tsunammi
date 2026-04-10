package openbook_v1

import (
	"encoding/binary"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
)

const serumDiscriminator = "serum"

func ParseMarketState(buf []byte) (*MarketState, error) {
	discriminator := string(buf[:5])

	switch discriminator {
	case serumDiscriminator:
		ms := MarketState{}
		decoder := ag_binary.NewBorshDecoder(buf[5:])

		err := ms.UnmarshalWithDecoder(decoder)
		if err != nil {
			return nil, err
		}

		return &ms, nil
	default:
		return nil, fmt.Errorf("unknown discriminator value: %s", discriminator)
	}
}

func ParseOpenOrders(buf []byte) (*OpenOrders, error) {
	discriminator := string(buf[:5])

	switch discriminator {
	case serumDiscriminator:
		orders := OpenOrders{}
		decoder := ag_binary.NewBorshDecoder(buf[5:])

		err := orders.UnmarshalWithDecoder(decoder)
		if err != nil {
			return nil, err
		}

		return &orders, nil
	default:
		return nil, fmt.Errorf("unknown discriminator value: %s", discriminator)
	}
}

type MarketState struct {
	AccountFlags uint64

	// 1
	OwnAddress ag_solanago.PublicKey

	// 5
	VaultSignerNonce uint64

	// 6
	CoinMint ag_solanago.PublicKey

	// 10
	PcMint ag_solanago.PublicKey

	// 14
	CoinVault ag_solanago.PublicKey

	// 18
	CoinDepositsTotal uint64

	// 19
	CoinFeesAccrued uint64

	// 20
	PcVault ag_solanago.PublicKey

	// 24
	PcDepositsTotal uint64

	// 25
	PcFeesAccrued uint64

	// 26
	PcDustThreshold uint64

	// 27
	ReqQ ag_solanago.PublicKey

	// 31
	EventQ ag_solanago.PublicKey

	// 35
	Bids ag_solanago.PublicKey

	// 39
	Asks ag_solanago.PublicKey

	// 43
	CoinLotSize uint64

	// 44
	PcLotSize uint64

	// 45
	FeeRateBps uint64

	// 46
	ReferrerRebatesAccrued uint64
}

func (obj *MarketState) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	ms := marketState{}

	err = ms.UnmarshalWithDecoder(decoder)
	if err != nil {
		return err
	}

	obj.AccountFlags = ms.AccountFlags
	obj.OwnAddress = uint64ArrayToPublicKey(ms.OwnAddress)
	obj.VaultSignerNonce = ms.VaultSignerNonce
	obj.CoinMint = uint64ArrayToPublicKey(ms.CoinMint)
	obj.PcMint = uint64ArrayToPublicKey(ms.PcMint)
	obj.CoinVault = uint64ArrayToPublicKey(ms.CoinVault)
	obj.CoinDepositsTotal = ms.CoinDepositsTotal
	obj.CoinFeesAccrued = ms.CoinFeesAccrued
	obj.PcVault = uint64ArrayToPublicKey(ms.PcVault)
	obj.PcDepositsTotal = ms.PcDepositsTotal
	obj.PcFeesAccrued = ms.PcFeesAccrued
	obj.PcDustThreshold = ms.PcDustThreshold
	obj.ReqQ = uint64ArrayToPublicKey(ms.ReqQ)
	obj.EventQ = uint64ArrayToPublicKey(ms.EventQ)
	obj.Bids = uint64ArrayToPublicKey(ms.Bids)
	obj.Asks = uint64ArrayToPublicKey(ms.Asks)
	obj.CoinLotSize = ms.CoinLotSize
	obj.PcLotSize = ms.PcLotSize
	obj.FeeRateBps = ms.FeeRateBps
	obj.ReferrerRebatesAccrued = ms.ReferrerRebatesAccrued

	return nil
}

func uint64ArrayToPublicKey(data [4]uint64) ag_solanago.PublicKey {

	buf := make([]byte, 32)

	binary.LittleEndian.PutUint64(buf[0:8], data[0])
	binary.LittleEndian.PutUint64(buf[8:16], data[1])
	binary.LittleEndian.PutUint64(buf[16:24], data[2])
	binary.LittleEndian.PutUint64(buf[24:32], data[3])

	return ag_solanago.PublicKeyFromBytes(buf)
}

// Ported from https://github.com/openbook-dex/program/blob/c85e56deeaead43abbc33b7301058838b9c5136d/dex/src/state.rs#L295
/*
pub struct MarketState {
// 0
pub account_flags: u64, // Initialized, Market

// 1
pub own_address: [u64; 4],

// 5
pub vault_signer_nonce: u64,
// 6
pub coin_mint: [u64; 4],
// 10
pub pc_mint: [u64; 4],

// 14
pub coin_vault: [u64; 4],
// 18
pub coin_deposits_total: u64,
// 19
pub coin_fees_accrued: u64,

// 20
pub pc_vault: [u64; 4],
// 24
pub pc_deposits_total: u64,
// 25
pub pc_fees_accrued: u64,

// 26
pub pc_dust_threshold: u64,

// 27
pub req_q: [u64; 4],
// 31
pub event_q: [u64; 4],

// 35
pub bids: [u64; 4],
// 39
pub asks: [u64; 4],

// 43
pub coin_lot_size: u64,
// 44
pub pc_lot_size: u64,

// 45
pub fee_rate_bps: u64,
// 46
pub referrer_rebates_accrued: u64,
}
*/
type marketState struct {
	// 0
	// Initialized, Market
	AccountFlags uint64

	// 1
	OwnAddress [4]uint64

	// 5
	VaultSignerNonce uint64

	// 6
	CoinMint [4]uint64

	// 10
	PcMint [4]uint64

	// 14
	CoinVault [4]uint64

	// 18
	CoinDepositsTotal uint64

	// 19
	CoinFeesAccrued uint64

	// 20
	PcVault [4]uint64

	// 24
	PcDepositsTotal uint64

	// 25
	PcFeesAccrued uint64

	// 26
	PcDustThreshold uint64

	// 27
	ReqQ [4]uint64

	// 31
	EventQ [4]uint64

	// 35
	Bids [4]uint64

	// 39
	Asks [4]uint64

	// 43
	CoinLotSize uint64

	// 44
	PcLotSize uint64

	// 45
	FeeRateBps uint64

	// 46
	ReferrerRebatesAccrued uint64
}

func (obj *marketState) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `AccountFlags` param:
	err = encoder.Encode(obj.AccountFlags)
	if err != nil {
		return err
	}
	// Serialize `OwnAddress` param:
	err = encoder.Encode(obj.OwnAddress)
	if err != nil {
		return err
	}
	// Serialize `VaultSignerNonce` param:
	err = encoder.Encode(obj.VaultSignerNonce)
	if err != nil {
		return err
	}
	// Serialize `CoinMint` param:
	err = encoder.Encode(obj.CoinMint)
	if err != nil {
		return err
	}
	// Serialize `PcMint` param:
	err = encoder.Encode(obj.PcMint)
	if err != nil {
		return err
	}
	// Serialize `CoinVault` param:
	err = encoder.Encode(obj.CoinVault)
	if err != nil {
		return err
	}
	// Serialize `CoinDepositsTotal` param:
	err = encoder.Encode(obj.CoinDepositsTotal)
	if err != nil {
		return err
	}
	// Serialize `CoinFeesAccrued` param:
	err = encoder.Encode(obj.CoinFeesAccrued)
	if err != nil {
		return err
	}
	// Serialize `PcVault` param:
	err = encoder.Encode(obj.PcVault)
	if err != nil {
		return err
	}
	// Serialize `PcDepositsTotal` param:
	err = encoder.Encode(obj.PcDepositsTotal)
	if err != nil {
		return err
	}
	// Serialize `PcFeesAccrued` param:
	err = encoder.Encode(obj.PcFeesAccrued)
	if err != nil {
		return err
	}
	// Serialize `PcDustThreshold` param:
	err = encoder.Encode(obj.PcDustThreshold)
	if err != nil {
		return err
	}
	// Serialize `ReqQ` param:
	err = encoder.Encode(obj.ReqQ)
	if err != nil {
		return err
	}
	// Serialize `EventQ` param:
	err = encoder.Encode(obj.EventQ)
	if err != nil {
		return err
	}
	// Serialize `Bids` param:
	err = encoder.Encode(obj.Bids)
	if err != nil {
		return err
	}
	// Serialize `Asks` param:
	err = encoder.Encode(obj.Asks)
	if err != nil {
		return err
	}
	// Serialize `CoinLotSize` param:
	err = encoder.Encode(obj.CoinLotSize)
	if err != nil {
		return err
	}
	// Serialize `PcLotSize` param:
	err = encoder.Encode(obj.PcLotSize)
	if err != nil {
		return err
	}
	// Serialize `FeeRateBps` param:
	err = encoder.Encode(obj.FeeRateBps)
	if err != nil {
		return err
	}
	// Serialize `ReferrerRebatesAccrued` param:
	err = encoder.Encode(obj.ReferrerRebatesAccrued)
	if err != nil {
		return err
	}
	return nil
}

func (obj *marketState) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `AccountFlags`:
	err = decoder.Decode(&obj.AccountFlags)
	if err != nil {
		return err
	}
	// Deserialize `OwnAddress`:
	err = decoder.Decode(&obj.OwnAddress)
	if err != nil {
		return err
	}
	// Deserialize `VaultSignerNonce`:
	err = decoder.Decode(&obj.VaultSignerNonce)
	if err != nil {
		return err
	}
	// Deserialize `CoinMint`:
	err = decoder.Decode(&obj.CoinMint)
	if err != nil {
		return err
	}
	// Deserialize `PcMint`:
	err = decoder.Decode(&obj.PcMint)
	if err != nil {
		return err
	}
	// Deserialize `CoinVault`:
	err = decoder.Decode(&obj.CoinVault)
	if err != nil {
		return err
	}
	// Deserialize `CoinDepositsTotal`:
	err = decoder.Decode(&obj.CoinDepositsTotal)
	if err != nil {
		return err
	}
	// Deserialize `CoinFeesAccrued`:
	err = decoder.Decode(&obj.CoinFeesAccrued)
	if err != nil {
		return err
	}
	// Deserialize `PcVault`:
	err = decoder.Decode(&obj.PcVault)
	if err != nil {
		return err
	}
	// Deserialize `PcDepositsTotal`:
	err = decoder.Decode(&obj.PcDepositsTotal)
	if err != nil {
		return err
	}
	// Deserialize `PcFeesAccrued`:
	err = decoder.Decode(&obj.PcFeesAccrued)
	if err != nil {
		return err
	}
	// Deserialize `PcDustThreshold`:
	err = decoder.Decode(&obj.PcDustThreshold)
	if err != nil {
		return err
	}
	// Deserialize `ReqQ`:
	err = decoder.Decode(&obj.ReqQ)
	if err != nil {
		return err
	}
	// Deserialize `EventQ`:
	err = decoder.Decode(&obj.EventQ)
	if err != nil {
		return err
	}
	// Deserialize `Bids`:
	err = decoder.Decode(&obj.Bids)
	if err != nil {
		return err
	}
	// Deserialize `Asks`:
	err = decoder.Decode(&obj.Asks)
	if err != nil {
		return err
	}
	// Deserialize `CoinLotSize`:
	err = decoder.Decode(&obj.CoinLotSize)
	if err != nil {
		return err
	}
	// Deserialize `PcLotSize`:
	err = decoder.Decode(&obj.PcLotSize)
	if err != nil {
		return err
	}
	// Deserialize `FeeRateBps`:
	err = decoder.Decode(&obj.FeeRateBps)
	if err != nil {
		return err
	}
	// Deserialize `ReferrerRebatesAccrued`:
	err = decoder.Decode(&obj.ReferrerRebatesAccrued)
	if err != nil {
		return err
	}
	return nil
}

type OpenOrders struct {
	// Initialized, OpenOrders
	AccountFlags    uint64
	Market          ag_solanago.PublicKey
	Owner           ag_solanago.PublicKey
	NativeCoinFree  uint64
	NativeCoinTotal uint64
	NativePcFree    uint64
	NativePcTotal   uint64
	FreeSlotBits    ag_binary.Uint128
	IsBidBits       ag_binary.Uint128
	Orders          [128]ag_binary.Uint128

	// Using Option<NonZeroU64> in a pod type requires nightly
	ClientOrderIds         [128]uint64
	ReferrerRebatesAccrued uint64
}

func (obj *OpenOrders) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	orders := openOrders{}

	err = orders.UnmarshalWithDecoder(decoder)

	if err != nil {
		return err
	}

	obj.AccountFlags = orders.AccountFlags
	obj.Market = uint64ArrayToPublicKey(orders.Market)
	obj.Owner = uint64ArrayToPublicKey(orders.Owner)
	obj.NativeCoinFree = orders.NativeCoinFree
	obj.NativeCoinTotal = orders.NativeCoinTotal
	obj.NativePcFree = orders.NativePcFree
	obj.NativePcTotal = orders.NativePcTotal
	obj.FreeSlotBits = orders.FreeSlotBits
	obj.IsBidBits = orders.IsBidBits
	obj.Orders = orders.Orders
	obj.ClientOrderIds = orders.ClientOrderIds
	obj.ReferrerRebatesAccrued = orders.ReferrerRebatesAccrued

	return nil
}

//Ported from https://github.com/openbook-dex/program/blob/c85e56deeaead43abbc33b7301058838b9c5136d/dex/src/state.rs#L589
/*
pub struct OpenOrders {
    pub account_flags: u64, // Initialized, OpenOrders
    pub market: [u64; 4],
    pub owner: [u64; 4],

    pub native_coin_free: u64,
    pub native_coin_total: u64,

    pub native_pc_free: u64,
    pub native_pc_total: u64,

    pub free_slot_bits: u128,
    pub is_bid_bits: u128,
    pub orders: [u128; 128],
    // Using Option<NonZeroU64> in a pod type requires nightly
    pub client_order_ids: [u64; 128],
    pub referrer_rebates_accrued: u64,
}
*/
type openOrders struct {
	// Initialized, OpenOrders
	AccountFlags    uint64
	Market          [4]uint64
	Owner           [4]uint64
	NativeCoinFree  uint64
	NativeCoinTotal uint64
	NativePcFree    uint64
	NativePcTotal   uint64
	FreeSlotBits    ag_binary.Uint128
	IsBidBits       ag_binary.Uint128
	Orders          [128]ag_binary.Uint128

	// Using Option<NonZeroU64> in a pod type requires nightly
	ClientOrderIds         [128]uint64
	ReferrerRebatesAccrued uint64
}

func (obj *openOrders) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `AccountFlags` param:
	err = encoder.Encode(obj.AccountFlags)
	if err != nil {
		return err
	}
	// Serialize `Market` param:
	err = encoder.Encode(obj.Market)
	if err != nil {
		return err
	}
	// Serialize `Owner` param:
	err = encoder.Encode(obj.Owner)
	if err != nil {
		return err
	}
	// Serialize `NativeCoinFree` param:
	err = encoder.Encode(obj.NativeCoinFree)
	if err != nil {
		return err
	}
	// Serialize `NativeCoinTotal` param:
	err = encoder.Encode(obj.NativeCoinTotal)
	if err != nil {
		return err
	}
	// Serialize `NativePcFree` param:
	err = encoder.Encode(obj.NativePcFree)
	if err != nil {
		return err
	}
	// Serialize `NativePcTotal` param:
	err = encoder.Encode(obj.NativePcTotal)
	if err != nil {
		return err
	}
	// Serialize `FreeSlotBits` param:
	err = encoder.Encode(obj.FreeSlotBits)
	if err != nil {
		return err
	}
	// Serialize `IsBidBits` param:
	err = encoder.Encode(obj.IsBidBits)
	if err != nil {
		return err
	}
	// Serialize `Orders` param:
	err = encoder.Encode(obj.Orders)
	if err != nil {
		return err
	}
	// Serialize `ClientOrderIds` param:
	err = encoder.Encode(obj.ClientOrderIds)
	if err != nil {
		return err
	}
	// Serialize `ReferrerRebatesAccrued` param:
	err = encoder.Encode(obj.ReferrerRebatesAccrued)
	if err != nil {
		return err
	}
	return nil
}

func (obj *openOrders) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `AccountFlags`:
	err = decoder.Decode(&obj.AccountFlags)
	if err != nil {
		return err
	}
	// Deserialize `Market`:
	err = decoder.Decode(&obj.Market)
	if err != nil {
		return err
	}
	// Deserialize `Owner`:
	err = decoder.Decode(&obj.Owner)
	if err != nil {
		return err
	}
	// Deserialize `NativeCoinFree`:
	err = decoder.Decode(&obj.NativeCoinFree)
	if err != nil {
		return err
	}
	// Deserialize `NativeCoinTotal`:
	err = decoder.Decode(&obj.NativeCoinTotal)
	if err != nil {
		return err
	}
	// Deserialize `NativePcFree`:
	err = decoder.Decode(&obj.NativePcFree)
	if err != nil {
		return err
	}
	// Deserialize `NativePcTotal`:
	err = decoder.Decode(&obj.NativePcTotal)
	if err != nil {
		return err
	}
	// Deserialize `FreeSlotBits`:
	err = decoder.Decode(&obj.FreeSlotBits)
	if err != nil {
		return err
	}
	// Deserialize `IsBidBits`:
	err = decoder.Decode(&obj.IsBidBits)
	if err != nil {
		return err
	}
	// Deserialize `Orders`:
	err = decoder.Decode(&obj.Orders)
	if err != nil {
		return err
	}
	// Deserialize `ClientOrderIds`:
	err = decoder.Decode(&obj.ClientOrderIds)
	if err != nil {
		return err
	}
	// Deserialize `ReferrerRebatesAccrued`:
	err = decoder.Decode(&obj.ReferrerRebatesAccrued)
	if err != nil {
		return err
	}
	return nil
}

// EventQueueHeader Ported from https://github.com/openbook-dex/program/blob/c85e56deeaead43abbc33b7301058838b9c5136d/dex/src/state.rs#L1103
/*
pub struct EventQueueHeader {
    account_flags: u64, // Initialized, EventQueue
    head: u64,
    count: u64,
    seq_num: u64,
}
*/
type EventQueueHeader struct {
	// Initialized, EventQueue
	AccountFlags uint64
	Head         uint64
	Count        uint64
	SeqNum       uint64
}

func (obj *EventQueueHeader) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `AccountFlags` param:
	err = encoder.Encode(obj.AccountFlags)
	if err != nil {
		return err
	}
	// Serialize `Head` param:
	err = encoder.Encode(obj.Head)
	if err != nil {
		return err
	}
	// Serialize `Count` param:
	err = encoder.Encode(obj.Count)
	if err != nil {
		return err
	}
	// Serialize `SeqNum` param:
	err = encoder.Encode(obj.SeqNum)
	if err != nil {
		return err
	}
	return nil
}

func (obj *EventQueueHeader) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `AccountFlags`:
	err = decoder.Decode(&obj.AccountFlags)
	if err != nil {
		return err
	}
	// Deserialize `Head`:
	err = decoder.Decode(&obj.Head)
	if err != nil {
		return err
	}
	// Deserialize `Count`:
	err = decoder.Decode(&obj.Count)
	if err != nil {
		return err
	}
	// Deserialize `SeqNum`:
	err = decoder.Decode(&obj.SeqNum)
	if err != nil {
		return err
	}
	return nil
}

type Event struct {
	EventFlags        uint8
	OwnerSlot         uint8
	FeeTier           uint8
	NativeQtyReleased uint64
	NativeQtyPaid     uint64
	NativeFeeOrRebate uint64
	OrderId           ag_binary.Uint128
	Owner             ag_solanago.PublicKey
	ClientOrderId     uint64
}

func (obj *Event) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	eventObj := event{}
	err = eventObj.UnmarshalWithDecoder(decoder)

	if err != nil {
		return err
	}
	obj.EventFlags = eventObj.EventFlags
	obj.OwnerSlot = eventObj.OwnerSlot
	obj.FeeTier = eventObj.FeeTier
	obj.NativeQtyReleased = eventObj.NativeQtyReleased
	obj.NativeQtyPaid = eventObj.NativeQtyPaid
	obj.NativeFeeOrRebate = eventObj.NativeFeeOrRebate
	obj.OrderId = eventObj.OrderId
	obj.Owner = uint64ArrayToPublicKey(eventObj.Owner)
	obj.ClientOrderId = eventObj.ClientOrderId

	return nil
}

// Ported from https://github.com/openbook-dex/program/blob/c85e56deeaead43abbc33b7301058838b9c5136d/dex/src/state.rs#L1171
/*
pub struct Event {
    event_flags: u8,
    owner_slot: u8,

    fee_tier: u8,

    _padding: [u8; 5],

    native_qty_released: u64,
    native_qty_paid: u64,
    native_fee_or_rebate: u64,

    order_id: u128,
    pub owner: [u64; 4],
    client_order_id: u64,
}
*/

type event struct {
	EventFlags        uint8
	OwnerSlot         uint8
	FeeTier           uint8
	Padding           [5]uint8
	NativeQtyReleased uint64
	NativeQtyPaid     uint64
	NativeFeeOrRebate uint64
	OrderId           ag_binary.Uint128
	Owner             [4]uint64
	ClientOrderId     uint64
}

func (obj *event) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `EventFlags` param:
	err = encoder.Encode(obj.EventFlags)
	if err != nil {
		return err
	}
	// Serialize `OwnerSlot` param:
	err = encoder.Encode(obj.OwnerSlot)
	if err != nil {
		return err
	}
	// Serialize `FeeTier` param:
	err = encoder.Encode(obj.FeeTier)
	if err != nil {
		return err
	}
	// Serialize `Padding` param:
	err = encoder.Encode(obj.Padding)
	if err != nil {
		return err
	}
	// Serialize `NativeQtyReleased` param:
	err = encoder.Encode(obj.NativeQtyReleased)
	if err != nil {
		return err
	}
	// Serialize `NativeQtyPaid` param:
	err = encoder.Encode(obj.NativeQtyPaid)
	if err != nil {
		return err
	}
	// Serialize `NativeFeeOrRebate` param:
	err = encoder.Encode(obj.NativeFeeOrRebate)
	if err != nil {
		return err
	}
	// Serialize `OrderId` param:
	err = encoder.Encode(obj.OrderId)
	if err != nil {
		return err
	}
	// Serialize `Owner` param:
	err = encoder.Encode(obj.Owner)
	if err != nil {
		return err
	}
	// Serialize `ClientOrderId` param:
	err = encoder.Encode(obj.ClientOrderId)
	if err != nil {
		return err
	}
	return nil
}

func (obj *event) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `EventFlags`:
	err = decoder.Decode(&obj.EventFlags)
	if err != nil {
		return err
	}
	// Deserialize `OwnerSlot`:
	err = decoder.Decode(&obj.OwnerSlot)
	if err != nil {
		return err
	}
	// Deserialize `FeeTier`:
	err = decoder.Decode(&obj.FeeTier)
	if err != nil {
		return err
	}
	// Deserialize `Padding`:
	err = decoder.Decode(&obj.Padding)
	if err != nil {
		return err
	}
	// Deserialize `NativeQtyReleased`:
	err = decoder.Decode(&obj.NativeQtyReleased)
	if err != nil {
		return err
	}
	// Deserialize `NativeQtyPaid`:
	err = decoder.Decode(&obj.NativeQtyPaid)
	if err != nil {
		return err
	}
	// Deserialize `NativeFeeOrRebate`:
	err = decoder.Decode(&obj.NativeFeeOrRebate)
	if err != nil {
		return err
	}
	// Deserialize `OrderId`:
	err = decoder.Decode(&obj.OrderId)
	if err != nil {
		return err
	}
	// Deserialize `Owner`:
	err = decoder.Decode(&obj.Owner)
	if err != nil {
		return err
	}
	// Deserialize `ClientOrderId`:
	err = decoder.Decode(&obj.ClientOrderId)
	if err != nil {
		return err
	}
	return nil
}
