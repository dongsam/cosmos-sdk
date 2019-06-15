// nolint
package tags

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ActionCompleteUnbonding    = "complete-unbonding"
	ActionCompleteRedelegation = "complete-redelegation"

	Action       = sdk.TagAction
	SrcValidator = sdk.TagSrcValidator
	DstValidator = sdk.TagDstValidator
	Delegator    = sdk.TagDelegator
	SrcDelegator = sdk.TagSrcDelegator
	DstDelegator = sdk.TagDstDelegator
	Moniker      = "moniker"
	Identity     = "identity"
	EndTime      = "end-time"
)
