package workers

import (
	"testing"

	ethCommon "github.com/arcology/3rd-party/eth/common"
)

func BenchmarkAggreSelector(b *testing.B) {
	listSize := 100000
	hashes := make([]*ethCommon.Hash, listSize)
	for i := 0; i < listSize; i++ {
		hash := ethCommon.BytesToHash([]byte{byte(i / 65536), byte(i / 256 % 256), byte(i % 256)})
		hashes[i] = &hash
	}

	// for i := 0; i < b.N; i++ {
	// 	aggreSelector := NewAggreSelector(1, "aggre-selector", 2)
	// 	//streamer := NewMockStreamer(b)
	// 	//aggreSelector.Init("aggre-selector", streamer)
	// 	aggreSelector.OnStart()

	// 	aggreSelector.OnMessageArrived([]*actor.Message{
	// 		{
	// 			Name: actor.MsgReaperCommand,
	// 			Data: &poolTypes.ReaperCommand{
	// 				NodeRole: &types.NodeRole{
	// 					Role: actor.MsgBlockRole_Validate,
	// 				},
	// 				Reapinglist: &types.ReapingList{
	// 					List:      hashes,
	// 					Timestamp: new(big.Int).SetInt64(1),
	// 				},
	// 			},
	// 		},
	// 	})

	// 	for j := 0; j < listSize; j++ {
	// 		aggreSelector.OnMessageArrived([]*actor.Message{
	// 			{
	// 				Name: actor.MsgMessager,
	// 				Data: &types.StandardMessage{
	// 					TxHash: *hashes[j],
	// 				},
	// 			},
	// 		})
	// 	}
	// }
}
