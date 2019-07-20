package types

import (
	"bytes"
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgCreateValidator{}
	_ sdk.Msg = &MsgEditValidator{}
	_ sdk.Msg = &MsgDelegate{}
	_ sdk.Msg = &MsgUndelegate{}
	_ sdk.Msg = &MsgBeginRedelegate{}
)

//______________________________________________________________________

// MsgCreateValidator - struct for bonding transactions
type MsgCreateValidator struct {
	Description       Description    `json:"description"`
	Commission        CommissionMsg  `json:"commission"`
	MinSelfDelegation sdk.Int        `json:"min_self_delegation"`
	DelegatorAddress  sdk.AccAddress `json:"delegator_address"`
	ValidatorAddress  sdk.ValAddress `json:"validator_address"`
	PubKey            crypto.PubKey  `json:"pubkey"`
	Value             sdk.Coin       `json:"value"`
	ShareTokenDenom   string         `json:"share_token_denom"`
}

type msgCreateValidatorJSON struct {
	Description       Description    `json:"description"`
	Commission        CommissionMsg  `json:"commission"`
	MinSelfDelegation sdk.Int        `json:"min_self_delegation"`
	DelegatorAddress  sdk.AccAddress `json:"delegator_address"`
	ValidatorAddress  sdk.ValAddress `json:"validator_address"`
	PubKey            string         `json:"pubkey"`
	Value             sdk.Coin       `json:"value"`
	ShareTokenDenom   string         `json:"share_token_denom"`
}

// Default way to create validator. Delegator address and validator address are the same
func NewMsgCreateValidator(
	valAddr sdk.ValAddress, pubKey crypto.PubKey, selfDelegation sdk.Coin,
	description Description, commission CommissionMsg, minSelfDelegation sdk.Int,
	shareTokenDenom string,
) MsgCreateValidator {

	return MsgCreateValidator{
		Description:       description,
		DelegatorAddress:  sdk.AccAddress(valAddr),
		ValidatorAddress:  valAddr,
		PubKey:            pubKey,
		Value:             selfDelegation,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
		ShareTokenDenom:   shareTokenDenom,
	}
}

//nolint
func (msg MsgCreateValidator) Route() string { return RouterKey }
func (msg MsgCreateValidator) Type() string  { return "create_validator" }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{msg.DelegatorAddress}

	if !bytes.Equal(msg.DelegatorAddress.Bytes(), msg.ValidatorAddress.Bytes()) {
		// if validator addr is not same as delegator addr, validator must sign
		// msg as well
		addrs = append(addrs, sdk.AccAddress(msg.ValidatorAddress))
	}
	return addrs
}

// MarshalJSON implements the json.Marshaler interface to provide custom JSON
// serialization of the MsgCreateValidator type.
func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
	return json.Marshal(msgCreateValidatorJSON{
		Description:       msg.Description,
		Commission:        msg.Commission,
		DelegatorAddress:  msg.DelegatorAddress,
		ValidatorAddress:  msg.ValidatorAddress,
		PubKey:            sdk.MustBech32ifyConsPub(msg.PubKey),
		Value:             msg.Value,
		MinSelfDelegation: msg.MinSelfDelegation,
		ShareTokenDenom:   msg.ShareTokenDenom,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface to provide custom
// JSON deserialization of the MsgCreateValidator type.
func (msg *MsgCreateValidator) UnmarshalJSON(bz []byte) error {
	var msgCreateValJSON msgCreateValidatorJSON
	if err := json.Unmarshal(bz, &msgCreateValJSON); err != nil {
		return err
	}

	msg.Description = msgCreateValJSON.Description
	msg.Commission = msgCreateValJSON.Commission
	msg.DelegatorAddress = msgCreateValJSON.DelegatorAddress
	msg.ValidatorAddress = msgCreateValJSON.ValidatorAddress
	var err error
	msg.PubKey, err = sdk.GetConsPubKeyBech32(msgCreateValJSON.PubKey)
	if err != nil {
		return err
	}
	msg.Value = msgCreateValJSON.Value
	msg.MinSelfDelegation = msgCreateValJSON.MinSelfDelegation
	msg.ShareTokenDenom = msgCreateValJSON.ShareTokenDenom

	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgCreateValidator) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgCreateValidator) ValidateBasic() sdk.Error {
	// note that unmarshaling from bech32 ensures either empty or valid
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if !sdk.AccAddress(msg.ValidatorAddress).Equals(msg.DelegatorAddress) {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	if msg.Value.Amount.LTE(sdk.ZeroInt()) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if msg.Commission == (CommissionMsg{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "commission must be included")
	}

	if msg.ShareTokenDenom == "" {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "share token denomination must be included.")
	}
	if !msg.MinSelfDelegation.GT(sdk.ZeroInt()) {
		return ErrMinSelfDelegationInvalid(DefaultCodespace)
	}
	if msg.Value.Amount.LT(msg.MinSelfDelegation) {
		return ErrSelfDelegationBelowMinimum(DefaultCodespace)
	}

	return nil
}

// MsgEditValidator - struct for editing a validator
type MsgEditValidator struct {
	Description
	ValidatorAddress sdk.ValAddress `json:"address"`

	// We pass a reference to the new commission rate and min self delegation as it's not mandatory to
	// update. If not updated, the deserialized rate will be zero with no way to
	// distinguish if an update was intended.
	//
	// REF: #2373
	CommissionRate    *sdk.Dec `json:"commission_rate"`
	MinSelfDelegation *sdk.Int `json:"min_self_delegation"`
}

func NewMsgEditValidator(valAddr sdk.ValAddress, description Description, newRate *sdk.Dec, newMinSelfDelegation *sdk.Int) MsgEditValidator {
	return MsgEditValidator{
		Description:       description,
		CommissionRate:    newRate,
		ValidatorAddress:  valAddr,
		MinSelfDelegation: newMinSelfDelegation,
	}
}

//nolint
func (msg MsgEditValidator) Route() string { return RouterKey }
func (msg MsgEditValidator) Type() string  { return "edit_validator" }
func (msg MsgEditValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddress)}
}

// get the bytes for the message signer to sign on
func (msg MsgEditValidator) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgEditValidator) ValidateBasic() sdk.Error {
	if msg.ValidatorAddress.Empty() {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "nil validator address")
	}

	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "transaction must include some information to modify")
	}

	if msg.MinSelfDelegation != nil && !(*msg.MinSelfDelegation).GT(sdk.ZeroInt()) {
		return ErrMinSelfDelegationInvalid(DefaultCodespace)
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(sdk.OneDec()) || msg.CommissionRate.LT(sdk.ZeroDec()) {
			return sdk.NewError(DefaultCodespace, CodeInvalidInput, "commission rate must be between 0 and 1, inclusive")
		}
	}

	return nil
}

// MsgDelegate - struct for bonding transactions
type MsgDelegate struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
	Amount           sdk.Coin       `json:"amount"`
}

func NewMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) MsgDelegate {
	return MsgDelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

// HACKATOM: Create New Message for Index Delegrations
// -----------------------------------------------------------------------------
type ValidatorPortion struct {
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
	Amount           sdk.Coin       `json:"amount"`
}

type MsgIndexDelegate struct {
	DelegatorAddress sdk.AccAddress     `json:"delegator_address"`
	Portions         []ValidatorPortion `json:"validator_portion"`
	Denomination     string             `json:"denomination"`
}

func (msg MsgIndexDelegate) Route() string { return RouterKey }
func (msg MsgIndexDelegate) Type() string  { return "index_delegate" }
func (msg MsgIndexDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

// get the bytes for the message signer to sign on
func (msg MsgIndexDelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgIndexDelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	for _, portion := range msg.Portions {
		if portion.ValidatorAddress.Empty() {
			return ErrNilValidatorAddr(DefaultCodespace)
		}
		if portion.Amount.Amount.LTE(sdk.ZeroInt()) {
			return ErrBadDelegationAmount(DefaultCodespace)
		}
	}
	return nil
}

// HACKATOM END
// -----------------------------------------------------------------------------

//nolint
func (msg MsgDelegate) Route() string { return RouterKey }
func (msg MsgDelegate) Type() string  { return "delegate" }
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

// get the bytes for the message signer to sign on
func (msg MsgDelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgDelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Amount.Amount.LTE(sdk.ZeroInt()) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	return nil
}

//______________________________________________________________________

// MsgDelegate - struct for bonding transactions
type MsgBeginRedelegate struct {
	DelegatorAddress    sdk.AccAddress `json:"delegator_address"`
	ValidatorSrcAddress sdk.ValAddress `json:"validator_src_address"`
	ValidatorDstAddress sdk.ValAddress `json:"validator_dst_address"`
	Amount              sdk.Coin       `json:"amount"`
}

func NewMsgBeginRedelegate(delAddr sdk.AccAddress, valSrcAddr,
	valDstAddr sdk.ValAddress, amount sdk.Coin) MsgBeginRedelegate {

	return MsgBeginRedelegate{
		DelegatorAddress:    delAddr,
		ValidatorSrcAddress: valSrcAddr,
		ValidatorDstAddress: valDstAddr,
		Amount:              amount,
	}
}

//nolint
func (msg MsgBeginRedelegate) Route() string { return RouterKey }
func (msg MsgBeginRedelegate) Type() string  { return "begin_redelegate" }
func (msg MsgBeginRedelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddress}
}

// get the bytes for the message signer to sign on
func (msg MsgBeginRedelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgBeginRedelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorSrcAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.ValidatorDstAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Amount.Amount.LTE(sdk.ZeroInt()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}

// MsgUndelegate - struct for unbonding transactions
type MsgUndelegate struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
	Amount           sdk.Coin       `json:"amount"`
}

func NewMsgUndelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) MsgUndelegate {
	return MsgUndelegate{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           amount,
	}
}

//nolint
func (msg MsgUndelegate) Route() string                { return RouterKey }
func (msg MsgUndelegate) Type() string                 { return "begin_unbonding" }
func (msg MsgUndelegate) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.DelegatorAddress} }

// get the bytes for the message signer to sign on
func (msg MsgUndelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgUndelegate) ValidateBasic() sdk.Error {
	if msg.DelegatorAddress.Empty() {
		return ErrNilDelegatorAddr(DefaultCodespace)
	}
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Amount.Amount.LTE(sdk.ZeroInt()) {
		return ErrBadSharesAmount(DefaultCodespace)
	}
	return nil
}
