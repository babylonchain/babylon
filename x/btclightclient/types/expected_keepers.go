package types

import (
	"context"
)

type BTCLightClientHooks interface {
	AfterBTCRollBack(ctx context.Context, headerInfo *BTCHeaderInfo)       // Must be called after the chain is rolled back
	AfterBTCRollForward(ctx context.Context, headerInfo *BTCHeaderInfo)    // Must be called after the chain is rolled forward
	AfterBTCHeaderInserted(ctx context.Context, headerInfo *BTCHeaderInfo) // Must be called after a header is inserted
}
