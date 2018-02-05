package state

import (
	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/events"

	"github.com/lino-network/lino/plugins/ibc"
	"github.com/lino-network/lino/types"
	ttx "github.com/lino-network/lino/types/tx"
)

// If the tx is invalid, a TMSP error will be returned.
func ExecTx(state *State, pgz *types.Plugins, tx ttx.Tx, isCheckTx bool, evc events.Fireable) abci.Result {
	chainID := state.GetChainID()

	// Exec tx
	switch tx := tx.(type) {
	case *ttx.SendTx:
		// Validate inputs and outputs, basic
		res := validateInputsBasic(tx.Inputs)
		if res.IsErr() {
			return res.PrependLog("in validateInputsBasic()")
		}
		res = validateOutputsBasic(tx.Outputs)
		if res.IsErr() {
			return res.PrependLog("in validateOutputsBasic()")
		}

		// Get inputs
		accounts, res := getInputs(state, tx.Inputs)
		if res.IsErr() {
			return res.PrependLog("in getInputs()")
		}

		// Get or make outputs.
		accounts, res = getOrMakeOutputs(state, accounts, tx.Outputs)
		if res.IsErr() {
			return res.PrependLog("in getOrMakeOutputs()")
		}

		// Validate inputs and outputs, advanced
		signBytes := tx.SignBytes(chainID)
		inTotal, res := validateInputsAdvanced(accounts, signBytes, tx.Inputs)
		if res.IsErr() {
			return res.PrependLog("in validateInputsAdvanced()")
		}
		outTotal := sumOutputs(tx.Outputs)
		outPlusFees := outTotal
		fees := types.Coins{tx.Fee}
		if fees.IsValid() { // TODO: fix coins.Plus()
			outPlusFees = outTotal.Plus(fees)
		}
		if !inTotal.IsEqual(outPlusFees) {
			return abci.ErrBaseInvalidOutput.AppendLog(cmn.Fmt("Input total (%v) != output total + fees (%v)", inTotal, outPlusFees))
		}

		// TODO: Fee validation for SendTx

		// Good! Adjust accounts
		adjustByInputs(state, accounts, tx.Inputs)
		adjustByOutputs(state, accounts, tx.Outputs, isCheckTx)

		/*
			// Fire events
			if !isCheckTx {
				if evc != nil {
					for _, i := range tx.Inputs {
						evc.FireEvent(types.EventStringAccInput(i.Address), types.EventDataTx{tx, nil, ""})
					}
					for _, o := range tx.Outputs {
						evc.FireEvent(types.EventStringAccOutput(o.Address), types.EventDataTx{tx, nil, ""})
					}
				}
			}
		*/

		return abci.NewResultOK(ttx.TxID(chainID, tx), "")

	case *ttx.AppTx:
		// Validate input, basic
		res := tx.Input.ValidateBasic()
		if res.IsErr() {
			return res
		}

		// Get input account
		inAcc := state.GetAccount(tx.Input.Address)
		if inAcc == nil {
			return abci.ErrBaseUnknownAddress
		}
		if !tx.Input.PubKey.Empty() {
			inAcc.PubKey = tx.Input.PubKey
		}

		// Validate input, advanced
		signBytes := tx.SignBytes(chainID)
		res = validateInputAdvanced(inAcc, signBytes, tx.Input)
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validateInputAdvanced failed on %X: %v", tx.Input.Address, res))
			return res.PrependLog("in validateInputAdvanced()")
		}
		if !tx.Input.Coins.IsGTE(types.Coins{tx.Fee}) {
			state.logger.Info(cmn.Fmt("Sender did not send enough to cover the fee %X", tx.Input.Address))
			return abci.ErrBaseInsufficientFunds.AppendLog(cmn.Fmt("input coins is %v, but fee is %v", tx.Input.Coins, types.Coins{tx.Fee}))
		}

		// Validate call address
		plugin := pgz.GetByName(tx.Name)
		if plugin == nil {
			return abci.ErrBaseUnknownAddress.AppendLog(
				cmn.Fmt("Unrecognized plugin name%v", tx.Name))
		}

		// Good!
		coins := tx.Input.Coins.Minus(types.Coins{tx.Fee})
		inAcc.Sequence += 1
		inAcc.Balance = inAcc.Balance.Minus(tx.Input.Coins)

		// If this is a CheckTx, stop now.
		if isCheckTx {
			state.SetAccount(tx.Input.Address, inAcc)
			return abci.OK
		}

		// Create inAcc checkpoint
		inAccCopy := inAcc.Copy()

		// Run the tx.
		cache := state.CacheWrap()
		cache.SetAccount(tx.Input.Address, inAcc)
		ctx := types.NewCallContext(tx.Input.Address, inAcc, coins)
		res = plugin.RunTx(cache, ctx, tx.Data)
		if res.IsOK() {
			cache.CacheSync()
			state.logger.Info("Successful execution")
			// Fire events
			/*
				if evc != nil {
					exception := ""
					if res.IsErr() {
						exception = res.Error()
					}
					evc.FireEvent(types.EventStringAccInput(tx.Input.Address), types.EventDataTx{tx, ret, exception})
					evc.FireEvent(types.EventStringAccOutput(tx.Address), types.EventDataTx{tx, ret, exception})
				}
			*/
		} else {
			state.logger.Info("AppTx failed", "error", res)
			// Just return the coins and return.
			inAccCopy.Balance = inAccCopy.Balance.Plus(coins)
			// But take the gas
			// TODO
			state.SetAccount(tx.Input.Address, inAccCopy)
		}
		return res

	case *ttx.PostTx:
		res := tx.ValidateBasic()
		if res.IsErr() {
			return res
		}
		if tx.Parent != nil && state.GetPost(tx.Parent) == nil {
			// TODO change to unknownpost
			return abci.ErrBaseUnknownAddress
		}
		// Get post author account
		acc := state.GetAccount(tx.Address)
		if acc == nil {
			acc = &types.Account{}
		}
		if !tx.PubKey.Empty() {
			acc.PubKey = tx.PubKey
		}

		signBytes := tx.SignBytes(chainID)
		res = validatePostAdvanced(acc, signBytes, *tx)
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validatePostAdvanced failed on %X: %v", tx.Address, res))
			return res.PrependLog("in validatePostAdvanced()")
		}
		acc.LastPost += 1
		post := &types.Post {
			Title: tx.Title,
			Content: tx.Content,
			Author: tx.Address,
			Parent: tx.Parent,
		}
		state.SetAccount(tx.Address, acc)
		state.SetPost(types.PostID(tx.Address, acc.LastPost), post)
		return abci.NewResultOK(ttx.TxID(chainID, tx), "")

	case *ttx.DonateTx:
		res := tx.ValidateBasic()
		if res.IsErr() {
			return res
		}
        acc := state.GetAccount(tx.Input.Address)
		if acc == nil {
			return abci.ErrBaseUnknownAddress
		}
		post := state.GetPost(tx.To)
		if post == nil {
			// TODO change to unknown post error
			return abci.ErrBaseUnknownAddress
		}
		if !tx.Input.PubKey.Empty() {
			acc.PubKey = tx.Input.PubKey
		}
		// Validate input, advanced
		signBytes := tx.SignBytes(chainID)
		res = validateInputAdvanced(acc, signBytes, tx.Input)
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validateInputAdvanced failed on %X: %v", tx.Input.Address, res))
			return res.PrependLog("in validateInputAdvanced()")
		}
		if !tx.Input.Coins.IsGTE(types.Coins{tx.Fee}) {
			state.logger.Info(cmn.Fmt("Sender did not send enough to cover the fee %X", tx.Input.Address))
			return abci.ErrBaseInsufficientFunds.AppendLog(cmn.Fmt("input coins is %v, but fee is %v", tx.Input.Coins, tx.Fee))
		}
		
		acc.Balance = acc.Balance.Minus(tx.Input.Coins)
		acc.Sequence += 1
		state.SetAccount(tx.Input.Address, acc)
		if isCheckTx {
			return abci.OK
		}

		outAcc := state.GetAccount(post.Author)
		if outAcc == nil {
			// TODO change to unknown post
			return abci.ErrBaseUnknownAddress
		}
		outCoin := tx.Input.Coins.Minus(types.Coins{tx.Fee})
		outAcc.Balance = outAcc.Balance.Plus(outCoin)
		state.SetAccount(post.Author, outAcc)
		return abci.NewResultOK(ttx.TxID(chainID, tx), "")

	case *ttx.LikeTx:
		res := tx.ValidateBasic()
		if res.IsErr() {
			return res
		}
		// Get post author account
		post := state.GetPost(tx.To)
		if post == nil {
			// TODO change to unknown Post
			return abci.ErrBaseUnknownAddress
		}
		account := state.GetAccount(tx.From)
		if account == nil {
			if tx.PubKey.Empty() {
				return abci.ErrBaseUnknownAddress
			}
			account = &types.Account{}
			account.PubKey = tx.PubKey
			state.SetAccount(tx.From, account)
		}
		signBytes := tx.SignBytes(chainID)
		res = validateLikeAdvanced(account, signBytes, *tx)
		if res.IsErr() {
			state.logger.Info(cmn.Fmt("validateLikeAdvanced failed on %X: %v", tx.From, res))
			return res.PrependLog("in validateLikeAdvanced()")
		}
		like := types.Like{
			From   : tx.From,
			To     : tx.To,
			Weight : tx.Weight,
		}
		state.AddLike(like)
		return abci.NewResultOK(ttx.TxID(chainID, tx), "")
	default:
		return abci.ErrBaseEncodingError.SetLog("Unknown tx type")
	}
}

//--------------------------------------------------------------------------------
// The accounts from the TxInputs must either already have
// crypto.PubKey.(type) != nil, (it must be known),
// or it must be specified in the TxInput.
func getInputs(state types.AccountGetter, ins []ttx.TxInput) (map[string]*types.Account, abci.Result) {
	accounts := map[string]*types.Account{}
	for _, in := range ins {
		// Account shouldn't be duplicated
		if _, ok := accounts[string(in.Address)]; ok {
			return nil, abci.ErrBaseDuplicateAddress
		}

		acc := state.GetAccount(in.Address)
		if acc == nil {
			return nil, abci.ErrBaseUnknownAddress
		}

		if !in.PubKey.Empty() {
			acc.PubKey = in.PubKey
		}
		accounts[string(in.Address)] = acc
	}
	return accounts, abci.OK
}

func getOrMakeOutputs(state types.AccountGetter, accounts map[string]*types.Account, outs []ttx.TxOutput) (map[string]*types.Account, abci.Result) {
	if accounts == nil {
		accounts = make(map[string]*types.Account)
	}

	for _, out := range outs {
		chain, outAddress, _ := out.ChainAndAddress() // already validated
		if chain != nil {
			// we dont need an account for the other chain.
			// we'll just create an outgoing ibc packet
			continue
		}
		// Account shouldn't be duplicated
		if _, ok := accounts[string(outAddress)]; ok {
			return nil, abci.ErrBaseDuplicateAddress
		}
		acc := state.GetAccount(outAddress)
		// output account may be nil (new)
		if acc == nil {
			// zero value is valid, empty account
			acc = &types.Account{}
		}
		accounts[string(outAddress)] = acc
	}
	return accounts, abci.OK
}

// Validate inputs basic structure
func validateInputsBasic(ins []ttx.TxInput) (res abci.Result) {
	for _, in := range ins {
		// Check TxInput basic
		if res := in.ValidateBasic(); res.IsErr() {
			return res
		}
	}
	return abci.OK
}

// Validate inputs and compute total amount of coins
func validateInputsAdvanced(accounts map[string]*types.Account, signBytes []byte, ins []ttx.TxInput) (total types.Coins, res abci.Result) {
	for _, in := range ins {
		acc := accounts[string(in.Address)]
		if acc == nil {
			cmn.PanicSanity("validateInputsAdvanced() expects account in accounts")
		}
		res = validateInputAdvanced(acc, signBytes, in)
		if res.IsErr() {
			return
		}
		// Good. Add amount to total
		total = total.Plus(in.Coins)
	}
	return total, abci.OK
}

func validateInputAdvanced(acc *types.Account, signBytes []byte, in ttx.TxInput) (res abci.Result) {
	// Check sequence/coins
	seq, balance := acc.Sequence, acc.Balance
	if seq+1 != in.Sequence {
		return abci.ErrBaseInvalidSequence.AppendLog(cmn.Fmt("Got %v, expected %v. (acc.seq=%v)", in.Sequence, seq+1, acc.Sequence))
	}
	// Check amount
	if !balance.IsGTE(in.Coins) {
		return abci.ErrBaseInsufficientFunds.AppendLog(cmn.Fmt("balance is %v, tried to send %v", balance, in.Coins))
	}
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, in.Signature) {
		return abci.ErrBaseInvalidSignature.AppendLog(cmn.Fmt("SignBytes: %X", signBytes))
	}
	return abci.OK
}

func validateOutputsBasic(outs []ttx.TxOutput) (res abci.Result) {
	for _, out := range outs {
		// Check TxOutput basic
		if res := out.ValidateBasic(); res.IsErr() {
			return res
		}
	}
	return abci.OK
}

func validatePostAdvanced(acc *types.Account, signBytes []byte, post ttx.PostTx) (res abci.Result) {
	// Check sequence
	seq := acc.LastPost
	if seq + 1 != post.Sequence {
		return abci.ErrBaseInvalidSequence.AppendLog(cmn.Fmt("Got %v, expected %v. (acc.seq=%v)", post.Sequence, seq+1, acc.LastPost))
	}
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, post.Signature) {
		return abci.ErrBaseInvalidSignature.AppendLog(cmn.Fmt("SignBytes: %X", signBytes))
	}
	return abci.OK
}

func validateLikeAdvanced(acc *types.Account, signBytes []byte, like ttx.LikeTx) (res abci.Result) {
	// Check signatures
	if !acc.PubKey.VerifyBytes(signBytes, like.Signature) {
		return abci.ErrBaseInvalidSignature.AppendLog(cmn.Fmt("SignBytes: %X", signBytes))
	}
	return abci.OK
}

func sumOutputs(outs []ttx.TxOutput) (total types.Coins) {
	for _, out := range outs {
		total = total.Plus(out.Coins)
	}
	return total
}

func adjustByInputs(state types.AccountSetter, accounts map[string]*types.Account, ins []ttx.TxInput) {
	for _, in := range ins {
		acc := accounts[string(in.Address)]
		if acc == nil {
			cmn.PanicSanity("adjustByInputs() expects account in accounts")
		}
		if !acc.Balance.IsGTE(in.Coins) {
			cmn.PanicSanity("adjustByInputs() expects sufficient funds")
		}
		acc.Balance = acc.Balance.Minus(in.Coins)
		acc.Sequence += 1
		state.SetAccount(in.Address, acc)
	}
}

func adjustByOutputs(state *State, accounts map[string]*types.Account, outs []ttx.TxOutput, isCheckTx bool) {
	for _, out := range outs {
		destChain, outAddress, _ := out.ChainAndAddress() // already validated
		if destChain != nil {
			payload := ibc.CoinsPayload{outAddress, out.Coins}
			ibc.SaveNewIBCPacket(state, state.GetChainID(), string(destChain), payload)
			continue
		}

		acc := accounts[string(outAddress)]
		if acc == nil {
			cmn.PanicSanity("adjustByOutputs() expects account in accounts")
		}
		acc.Balance = acc.Balance.Plus(out.Coins)
		if !isCheckTx {
			state.SetAccount(outAddress, acc)
		}
	}
}