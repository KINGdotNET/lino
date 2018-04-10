package global

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/global/model"
	"github.com/lino-network/lino/types"
	"github.com/lino-network/lino/genesis"
)

// GlobalManager encapsulates all basic struct
type GlobalManager struct {
	globalStorage *model.GlobalStorage `json:"global_manager"`
}

// NewGlobalManager return the global proxy pointer
func NewGlobalManager(key sdk.StoreKey) *GlobalManager {
	return &GlobalManager{
		globalStorage: model.NewGlobalStorage(key),
	}
}

func (gm *GlobalManager) InitGlobalManager(ctx sdk.Context, state genesis.GlobalState) error {
	return gm.globalStorage.InitGlobalState(ctx, state)
}

func (gm *GlobalManager) registerEventAtHeight(ctx sdk.Context, height int64, event types.Event) sdk.Error {
	eventList, _ := gm.globalStorage.GetHeightEventList(ctx, height)
	if eventList == nil {
		eventList = &types.HeightEventList{Events: []types.Event{}}
	}
	eventList.Events = append(eventList.Events, event)
	if err := gm.globalStorage.SetHeightEventList(ctx, height, eventList); err != nil {
		return ErrGlobalManagerRegisterEventAtHeight(height).TraceCause(err, "")
	}
	return nil
}

func (gm *GlobalManager) registerEventAtTime(ctx sdk.Context, unixTime int64, event types.Event) sdk.Error {
	eventList, _ := gm.globalStorage.GetTimeEventList(ctx, unixTime)
	if eventList == nil {
		eventList = &types.TimeEventList{Events: []types.Event{}}
	}
	eventList.Events = append(eventList.Events, event)
	if err := gm.globalStorage.SetTimeEventList(ctx, unixTime, eventList); err != nil {
		return ErrGlobalManagerRegisterEventAtTime(unixTime).TraceCause(err, "")
	}
	return nil
}

func (gm *GlobalManager) GetHeightEventListAtHeight(ctx sdk.Context, height int64) (*types.HeightEventList) {
	eventList, _ := gm.globalStorage.GetHeightEventList(ctx, height)
	return eventList
}

func (gm *GlobalManager) RemoveHeightEventList(ctx sdk.Context, height int64) sdk.Error {
	return gm.globalStorage.RemoveHeightEventList(ctx, height)
}

func (gm *GlobalManager) GetTimeEventListAtTime(ctx sdk.Context, unixTime int64) (*types.TimeEventList) {
	eventList, _ := gm.globalStorage.GetTimeEventList(ctx, unixTime)
	return eventList
}

func (gm *GlobalManager) RemoveTimeEventList(ctx sdk.Context, unixTime int64) sdk.Error {
	return gm.globalStorage.RemoveTimeEventList(ctx, unixTime)
}


func (gm *GlobalManager) GetConsumptionFrictionRate(ctx sdk.Context) (sdk.Rat, sdk.Error) {
	consumptionMeta, err := gm.globalStorage.GetConsumptionMeta(ctx)
	if err != nil {
		return sdk.Rat{}, err
	}
	return consumptionMeta.ConsumptionFrictionRate, nil
}

// register reward calculation event at 7 days later
func (gm *GlobalManager) RegisterContentRewardEvent(ctx sdk.Context, event types.Event) sdk.Error {
	consumptionMeta, err := gm.globalStorage.GetConsumptionMeta(ctx)
	if err != nil {
		return err
	}
	if err := gm.registerEventAtTime(ctx, ctx.BlockHeader().Time+(consumptionMeta.FreezingPeriodHr*3600), event); err != nil {
		return err
	}
	return nil
}

// put a friction of user consumption to reward pool
func (gm *GlobalManager) AddConsumptionFrictionToRewardPool(ctx sdk.Context, coin types.Coin) sdk.Error {
	// skip micro micro payment (etc: 0.0001 LNO)
	if coin.IsZero() {
		return nil
	}

	consumptionMeta, err := gm.globalStorage.GetConsumptionMeta(ctx)
	if err != nil {
		return ErrAddConsumptionFrictionToRewardPool().TraceCause(err, "")
	}

	// reward pool consists of a small friction of user consumption and hourly content creator reward
	// consumption window will be used to calculate the percentage of reward to claim for this consumption
	consumptionMeta.ConsumptionRewardPool = consumptionMeta.ConsumptionRewardPool.Plus(coin)
	consumptionMeta.ConsumptionWindow = consumptionMeta.ConsumptionWindow.Plus(coin)

	if err := gm.globalStorage.SetConsumptionMeta(ctx, consumptionMeta); err != nil {
		return ErrAddConsumptionFrictionToRewardPool().TraceCause(err, "")
	}
	return nil
}

// after 7 days, one consumption needs to claim its reward from consumption reward pool
func (gm *GlobalManager) GetRewardAndPopFromWindow(ctx sdk.Context, coin types.Coin) (types.Coin, sdk.Error) {
	if coin.IsZero() {
		return types.NewCoin(0), nil
	}

	consumptionMeta, err := gm.globalStorage.GetConsumptionMeta(ctx)
	if err != nil {
		return types.NewCoin(0), ErrGetRewardAndPopFromWindow().TraceCause(err, "")
	}

	// reward = (consumption reward pool) * ((this consumption) / (total consumption in 7 days window))
	reward := types.RatToCoin(consumptionMeta.ConsumptionRewardPool.ToRat().
		Mul(coin.ToRat().Quo(consumptionMeta.ConsumptionWindow.ToRat())))

	consumptionMeta.ConsumptionRewardPool = consumptionMeta.ConsumptionRewardPool.Minus(reward)
	consumptionMeta.ConsumptionWindow = consumptionMeta.ConsumptionWindow.Minus(coin)

	if err := gm.globalStorage.SetConsumptionMeta(ctx, consumptionMeta); err != nil {
		return types.NewCoin(0), ErrGetRewardAndPopFromWindow().TraceCause(err, "")
	}
	return reward, nil
}

// add consumption to global meta, which is used to compute GDP
func (gm *GlobalManager) AddConsumption(ctx sdk.Context, coin types.Coin) sdk.Error {
	globalMeta, err := gm.globalStorage.GetGlobalMeta(ctx)
	if err != nil {
		return err
	}
	globalMeta.CumulativeConsumption = globalMeta.CumulativeConsumption.Plus(coin)

	if err := gm.globalStorage.SetGlobalMeta(ctx, globalMeta); err != nil {
		return err
	}
	return nil
}
