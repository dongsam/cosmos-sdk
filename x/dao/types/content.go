package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Constants pertaining to a Content object
//const (
//	MaxDescriptionLength int = 5000
//	MaxTitleLength       int = 140
//)

// Content defines an interface that a proposal must implement. It contains
// information such as the title and description along with the type and routing
// information for the appropriate handler to process the proposal. Content can
// have additional fields, which will handled by a proposal's Handler.
type Content interface {
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() sdk.Error
	String() string
}

// Handler defines a function that handles a proposal after it has passed the
// governance process.
type Handler func(ctx sdk.Context, content Content) sdk.Error

// ValidateAbstract validates a proposal's abstract contents returning an error
// if invalid.
func ValidateAbstract(codespace sdk.CodespaceType, c Content) sdk.Error {
	// TBD
	return nil
}
