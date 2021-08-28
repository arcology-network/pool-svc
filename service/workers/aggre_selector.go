package workers

import (
	ethCommon "github.com/arcology/3rd-party/eth/common"
	"github.com/arcology/common-lib/types"
	"github.com/arcology/component-lib/actor"
	"go.uber.org/zap"

	"github.com/arcology/component-lib/aggregator/aggregator"

	"github.com/arcology/component-lib/log"
	poolTypes "github.com/arcology/pool-svc/service/types"
)

type AggreSelector struct {
	actor.WorkerThread
	aggregator *aggregator.Aggregator
	maxReap    int
}

//return a Subscriber struct
func NewAggreSelector(concurrency int, groupid string, maxReap int) *AggreSelector {
	agg := AggreSelector{}
	agg.Set(concurrency, groupid)
	agg.aggregator = aggregator.NewAggregator()
	agg.maxReap = maxReap
	return &agg
}

func (a *AggreSelector) OnStart() {
}

func (a *AggreSelector) OnMessageArrived(msgs []*actor.Message) error {
	switch msgs[0].Name {
	case actor.MsgClearCommand:
		remainingQuantity := a.aggregator.OnClearInfoReceived()
		a.AddLog(log.LogLevel_Debug, "pool AggreSelector clear pool", zap.Int("remainingQuantity", remainingQuantity))
		a.MsgBroker.Send(actor.MsgClearCompleted, "")
	case actor.MsgStartReapCommand, actor.MsgInitReapCommand:
		a.AddLog(log.LogLevel_Debug, "pool AggreSelector reap Max ", zap.Int("nums", a.maxReap))
		_, result := a.aggregator.Reap(a.maxReap)
		a.SendMsg(result, true)
	case actor.MsgReapinglist:
		reapinglist := msgs[0].Data.(*types.ReapingList)
		a.AddLog(log.LogLevel_Debug, "pool AggreSelector reapingList ", zap.Int("nums", len(reapinglist.List)))
		result, _ := a.aggregator.OnListReceived(reapinglist)
		a.SendMsg(result, false)
	case actor.MsgMessager:
		messages := msgs[0].Data.([]*types.StandardMessage)
		datas := types.StandardMessages(messages).EncodeToBytes()

		a.AddLog(log.LogLevel_Debug, "received msg MsgMessager------------------------", zap.Int("messages size", len(messages)))

		for i := range messages {

			msg := poolTypes.SavingStandardMessage{
				Msg:     messages[i],
				RawData: datas[i],
			}
			result := a.aggregator.OnDataReceived(messages[i].TxHash, &msg)

			a.SendMsg(result, false)
		}
	}
	return nil
}
func (a *AggreSelector) SendMsg(selectedData *[]*interface{}, isProposer bool) {
	if selectedData != nil {
		messagerRawDatas := make([][]byte, len(*selectedData))
		txs := make([][]byte, len(*selectedData))
		hashlist := make([]*ethCommon.Hash, len(*selectedData))
		for i, msg := range *selectedData {
			savingStandardMessage := (*msg).(*poolTypes.SavingStandardMessage)
			messagerRawDatas[i] = savingStandardMessage.RawData
			txs[i] = savingStandardMessage.Msg.TxRawData
			hashlist[i] = &savingStandardMessage.Msg.TxHash
		}

		a.AddLog(log.LogLevel_Debug, "pool reapTxs end", zap.Int("txs", len(messagerRawDatas)))

		if isProposer {
			a.MsgBroker.Send(actor.MsgMetaBlock, &types.MetaBlock{
				//Txs: txs,
				Txs:      [][]byte{},
				Hashlist: hashlist,
			})
		} else {
			a.MsgBroker.Send(actor.MsgMessagersReaped, types.SendingStandardMessages{
				Data: messagerRawDatas,
			})
			a.MsgBroker.Send(actor.MsgSelectedTx, types.Txs{Data: txs})
			a.MsgBroker.Send(actor.MsgClearCommand, "")

		}
	}
}
