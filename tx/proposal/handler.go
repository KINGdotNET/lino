package proposal

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	acc "github.com/lino-network/lino/tx/account"
	"github.com/lino-network/lino/tx/global"
)

func NewHandler(am acc.AccountManager, pm ProposalManager, gm global.GlobalManager) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case ChangeGlobalAllocationMsg:
			return handleChangeGlobalAllocationMsg(ctx, am, pm, gm, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized vote Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleChangeGlobalAllocationMsg(
	ctx sdk.Context, am acc.AccountManager, pm ProposalManager, gm global.GlobalManager, msg ChangeGlobalAllocationMsg) sdk.Result {
	if !am.IsAccountExist(ctx, msg.Creator) {
		return ErrUsernameNotFound().Result()
	}

	// TODO add deposit logic
	if _, err := pm.AddProposal(ctx, msg.Creator, msg.Description, gm); err != nil {
		return err.Result()
	}
	//  set a time event to decide the proposal in 7 days
	// if err := pm.CreateDecideProposalEvent(ctx, gm); err != nil {
	// 	return err.Result()
	// }
	return sdk.Result{}
}
