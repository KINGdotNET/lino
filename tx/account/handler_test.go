package account

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/types"
	"github.com/stretchr/testify/assert"
)

var (
	l0    = types.LNO(sdk.NewRat(0))
	l100  = types.LNO(sdk.NewRat(100))
	l200  = types.LNO(sdk.NewRat(200))
	l1600 = types.LNO(sdk.NewRat(1600))
	l1800 = types.LNO(sdk.NewRat(1800))
	l1900 = types.LNO(sdk.NewRat(1900))
	l2000 = types.LNO(sdk.NewRat(2000))
	c0    = types.Coin{0}
	c100  = types.Coin{100 * types.Decimals}
	c200  = types.Coin{200 * types.Decimals}
	c1600 = types.Coin{1600 * types.Decimals}
	c1800 = types.Coin{1800 * types.Decimals}
	c1900 = types.Coin{1900 * types.Decimals}
	c2000 = types.Coin{2000 * types.Decimals}
)

func TestFollow(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create two test users
	createTestAccount(ctx, am, "user1")
	createTestAccount(ctx, am, "user2")

	// let user1 follows user2
	msg := NewFollowMsg("user1", "user2")
	result := handler(ctx, msg)
	assert.Equal(t, sdk.Result{}, result)

	// check user1 in the user2's follower list
	assert.Equal(t, true, am.IsMyFollowing(ctx, types.AccountKey("user1"), types.AccountKey("user2")))

	// check user2 in the user1's following list
	assert.Equal(t, true, am.IsMyFollower(ctx, types.AccountKey("user2"), types.AccountKey("user1")))
}

func TestFollowUserNotExist(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create test user
	createTestAccount(ctx, am, "user1")

	// let user2(not exists) follows user1
	msg := NewFollowMsg("user2", "user1")
	result := handler(ctx, msg)

	assert.Equal(t, result, ErrUsernameNotFound().Result())
	assert.Equal(t, false, am.IsMyFollower(ctx, types.AccountKey("user1"), types.AccountKey("user2")))

	// let user1 follows user3(not exists)
	msg = NewFollowMsg("user1", "user3")
	result = handler(ctx, msg)
	assert.Equal(t, result, ErrUsernameNotFound().Result())
	assert.Equal(t, false, am.IsMyFollowing(ctx, types.AccountKey("user1"), types.AccountKey("user3")))
}

func TestFollowAgain(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create two test users
	createTestAccount(ctx, am, "user1")
	createTestAccount(ctx, am, "user2")

	// let user1 follows user2 twice
	msg := NewFollowMsg("user1", "user2")
	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	msg = NewFollowMsg("user1", "user2")
	result = handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	// check user1 is user2's only follower
	assert.Equal(t, true, am.IsMyFollower(ctx, types.AccountKey("user2"), types.AccountKey("user1")))

	// check user2 is the only one in the user1's following list
	assert.Equal(t, true, am.IsMyFollowing(ctx, types.AccountKey("user1"), types.AccountKey("user2")))
}

func TestUnfollow(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create two test users
	createTestAccount(ctx, am, "user1")
	createTestAccount(ctx, am, "user2")

	// let user1 follows user2
	msg := NewFollowMsg("user1", "user2")
	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	// let user1 unfollows user2
	msg2 := NewUnfollowMsg("user1", "user2")
	result = handler(ctx, msg2)
	assert.Equal(t, result, sdk.Result{})

	// check user1 is not in the user2's follower list
	assert.Equal(t, false, am.IsMyFollower(ctx, types.AccountKey("user2"), types.AccountKey("user1")))

	// check user2 is not in the user1's following list
	assert.Equal(t, false, am.IsMyFollowing(ctx, types.AccountKey("user1"), types.AccountKey("user2")))
}

func TestUnfollowUserNotExist(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)
	// create test user
	createTestAccount(ctx, am, "user1")

	// let user2(not exists) unfollows user1
	msg := NewUnfollowMsg("user2", "user1")
	result := handler(ctx, msg)
	assert.Equal(t, result, ErrUsernameNotFound().Result())

	// let user1 unfollows user3(not exists)
	msg = NewUnfollowMsg("user1", "user3")
	result = handler(ctx, msg)
	assert.Equal(t, result, ErrUsernameNotFound().Result())
}

func TestInvalidUnfollow(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)
	// create test user
	createTestAccount(ctx, am, "user1")
	createTestAccount(ctx, am, "user2")
	createTestAccount(ctx, am, "user3")

	// let user1 follows user2
	msg := NewFollowMsg("user1", "user2")
	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	// let user3 unfollows user1 and user2 unfollows user3 (invalid)
	//this won't make any changes
	msg2 := NewUnfollowMsg("user3", "user1")
	result = handler(ctx, msg2)
	assert.Equal(t, result, sdk.Result{})

	msg3 := NewUnfollowMsg("user2", "user3")
	result = handler(ctx, msg3)
	assert.Equal(t, result, sdk.Result{})

	// check user1 in the user2's follower list
	assert.Equal(t, true, am.IsMyFollower(ctx, types.AccountKey("user2"), types.AccountKey("user1")))

	// check user2 in the user1's following list
	assert.Equal(t, true, am.IsMyFollowing(ctx, types.AccountKey("user1"), types.AccountKey("user2")))

}

func TestTransferNormal(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create two test users
	createTestAccount(ctx, am, "user1")
	createTestAccount(ctx, am, "user2")

	am.AddCoin(ctx, types.AccountKey("user1"), c2000)

	memo := []byte("This is a memo!")

	// let user1 transfers 200 to user2 (by username)
	msg := NewTransferMsg("user1", l200, memo, TransferToUser("user2"))
	result := handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	acc1Balance, _ := am.GetBankBalance(ctx, types.AccountKey("user1"))
	acc2Balance, _ := am.GetBankBalance(ctx, types.AccountKey("user2"))
	assert.Equal(t, c1800, acc1Balance)
	assert.Equal(t, true, acc2Balance.IsEqual(c200))

	acc2Addr, _ := am.GetBankAddress(ctx, types.AccountKey("user2"))
	msg = NewTransferMsg("user1", l1600, memo, TransferToUser("user2"), TransferToAddr(acc2Addr))
	result = handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	acc1Balance, _ = am.GetBankBalance(ctx, types.AccountKey("user1"))
	acc2Balance, _ = am.GetBankBalance(ctx, types.AccountKey("user2"))

	assert.Equal(t, true, acc1Balance.IsEqual(c200))
	assert.Equal(t, true, acc2Balance.IsEqual(c1800))

	msg = NewTransferMsg("user1", l100, memo, TransferToAddr(acc2Addr))
	result = handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	acc1Balance, _ = am.GetBankBalance(ctx, types.AccountKey("user1"))
	acc2Balance, _ = am.GetBankBalance(ctx, types.AccountKey("user2"))

	assert.Equal(t, true, acc1Balance.IsEqual(c100))
	assert.Equal(t, true, acc2Balance.IsEqual(c1900))

	randomAddr := sdk.Address("sdajsdbiqwbdiub")
	msg = NewTransferMsg("user1", l100, memo, TransferToAddr(randomAddr))
	result = handler(ctx, msg)
	assert.Equal(t, result, sdk.Result{})

	acc1Balance, _ = am.GetBankBalance(ctx, types.AccountKey("user1"))

	assert.Equal(t, true, acc1Balance.IsEqual(c0))

}

func TestSenderCoinNotEnough(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create two test users
	createTestAccount(ctx, am, "user1")
	createTestAccount(ctx, am, "user2")

	am.AddCoin(ctx, types.AccountKey("user1"), c1600)

	memo := []byte("This is a memo!")

	// let user1 transfers 2000 to user2
	msg := NewTransferMsg("user1", l2000, memo, TransferToUser("user2"))
	result := handler(ctx, msg)
	assert.Equal(t, ErrAccountCoinNotEnough().Result(), result)

	acc1Balance, _ := am.GetBankBalance(ctx, types.AccountKey("user1"))
	assert.Equal(t, true, acc1Balance.IsEqual(c1600))
}

func TestUsernameAddressMismatch(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create two test users
	createTestAccount(ctx, am, "user1")
	createTestAccount(ctx, am, "user2")

	am.AddCoin(ctx, types.AccountKey("user1"), c2000)
	am.AddCoin(ctx, types.AccountKey("user2"), c2000)

	memo := []byte("This is a memo!")
	randomAddr := sdk.Address("dqwdnqwdbnqwkjd")

	// let user1 transfers 2000 to user2 (provide both name and address)
	msg := NewTransferMsg("user1", l2000, memo, TransferToUser("user2"), TransferToAddr(randomAddr))
	result := handler(ctx, msg)
	assert.Equal(t, ErrTransferHandler(msg.Sender).Result(), result)
}

func TestReceiverUsernameIncorrect(t *testing.T) {
	ctx, am := setupTest(t, 1)
	handler := NewHandler(*am)

	// create two test users
	createTestAccount(ctx, am, "user1")
	am.AddCoin(ctx, types.AccountKey("user1"), c2000)

	memo := []byte("This is a memo!")

	// let user1 transfers 2000 to a random user
	msg := NewTransferMsg("user1", l2000, memo, TransferToUser("dnqwondqowindow"))
	result := handler(ctx, msg)
	assert.Equal(t, ErrTransferHandler(msg.Sender).Result().Code, result.Code)
}
