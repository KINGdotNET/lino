package post

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/types"
)

func ErrGetRedistributionSplitRate(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("post manager got %v redistribution split rate failed", permLink))
}

func ErrGetCreatedTime(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("post manager got %v created time failed", permLink))
}

func ErrGetRootSourcePost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("post manager got %v root source post failed", permLink))
}

func ErrSetRootSourcePost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("post manager set %v root source post failed", permLink))
}

func ErrCreatePost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("post manager created post %v failed", permLink))
}

func ErrCreatePostSourceInvalid(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("post manager created post %v failed, source post is invalid", permLink))
}

func ErrPostExist(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("post %v already exist", permLink))
}

func ErrAddOrUpdateLikeToPost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("add or update like to post %v failed", permLink))
}

func ErrAddOrUpdateReportOrUpvoteToPost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("add or update report or upvote to post %v failed", permLink))
}

func ErrRevokeReportOrUpvoteToPost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("revoke report or upvote to post %v failed", permLink))
}

func ErrAddDonation(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("add donation to post %v failed", permLink))
}

func ErrGetPenaltyScore(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostManagerError, fmt.Sprintf("get post %v penalty score failed", permLink))
}

func ErrCreatePostAuthorNotFound(author types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("create post author %v not found", author))
}

func ErrCreateExistPost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("create post failed, post %v already exist", permLink))
}

func ErrLikePostUserNotFound(user types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("like post failed, user %v not found", user))
}

func ErrLikeNonExistPost(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("like post failed, target post %v not found", permLink))
}

func ErrDonateFailed(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("donate to post %v failed", permLink))
}

func ErrDonateUserNotFound(user types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("donation failed, user %v not found", user))
}

func ErrDonateAuthorNotFound(permLink types.PermLink, author types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("donation failed, post %v author %v not found", permLink, author))
}

func ErrDonatePostDoesntExist(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("donate to post %v failed, post doesn't exist", permLink))
}

func ErrReportFailed(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("report to post %v failed", permLink))
}

func ErrReportUserNotFound(user types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("report failed, user %v not found", user))
}

func ErrReportAuthorNotFound(permLink types.PermLink, author types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("report failed, post %v author %v not found", permLink, author))
}

func ErrReportPostDoesntExist(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("report to post %v failed, post doesn't exist", permLink))
}

func ErrUpvoteUserNotFound(user types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("upvote failed, user %v not found", user))
}

func ErrUpvoteAuthorNotFound(permLink types.PermLink, author types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("upvote failed, post %v author %v not found", permLink, author))
}

func ErrUpvotePostDoesntExist(permLink types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostHandlerError, fmt.Sprintf("upvote to post %v failed, post doesn't exist", permLink))
}

func ErrPostCreateNoPostID() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Create with empty post id"))
}

func ErrPostCreateNoAuthor() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Create with empty author"))
}

func ErrCommentAndRepostError() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Post can't be comment and repost at the same time"))
}

func ErrCommentInvalidParent(parentPostKey types.PermLink) sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("comment post parent %v doesn't exist", parentPostKey))
}

func ErrPostLikeNoUsername() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Like needs have username"))
}

func ErrPostLikeWeightOverflow(weight int64) sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Like weight overflow: %v", weight))
}

func ErrPostLikeInvalidTarget() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Like target post invalid"))
}

func ErrPostReportOrUpvoteNoUsername() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("report or upvote needs have username"))
}

func ErrPostReportOrUpvoteInvalidTarget() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("report or upvote target post invalid"))
}

func ErrPostTitleExceedMaxLength() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Post title exceeds max length limitation"))
}

func ErrPostContentExceedMaxLength() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Post content exceeds max length limitation"))
}

func ErrPostRedistributionSplitRate() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Post redistribution rate invalid"))
}

func ErrPostDonateNoUsername() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Donate needs have username"))
}

func ErrPostDonateInvalidTarget() sdk.Error {
	return sdk.NewError(types.CodePostMsgError, fmt.Sprintf("Donate target post invalid"))
}
