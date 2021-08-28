package types

import (
	"github.com/arcology/common-lib/types"
)

type SavingStandardMessage struct {
	Msg     *types.StandardMessage
	RawData []byte
}
