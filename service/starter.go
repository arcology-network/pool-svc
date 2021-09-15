package service

import (
	cmn "github.com/arcology-network/3rd-party/tm/common"
	"github.com/arcology-network/component-lib/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start pool service Daemon",
	RunE:  startCmd,
}

func init() {
	flags := StartCmd.Flags()

	flags.String("mqaddr", "localhost:9092", "host:port of kafka")
	flags.String("mqaddr2", "localhost:9092", "host:port of kafka")

	flags.String("msgexch", "msgexch", "topic for receive msg exchange")
	flags.Int("concurrency", 4, "num of threads")
	flags.Int("rpm", 0, "max num of reap txs,0-not limit")
	flags.String("logcfg", "./log.toml", "log conf path")

	flags.String("chkd-message", "chkd-message", "topic for received message ")
	flags.String("reaping-list", "reaping-list", "topic of send receive reaping list")

	flags.Int("nidx", 0, "node index in cluster")
	flags.String("nname", "node1", "node name in cluster")

	flags.String("meta-block", "meta-block", "topic of send meta block ")
	flags.String("selected-txs", "selected-txs", "topic of send selected txs ")
	flags.String("selected-msgs", "selected-msgs", "topic of send selected msgs ")

}

func startCmd(cmd *cobra.Command, args []string) error {
	log.InitLog("pool.log", viper.GetString("logcfg"), "pool", viper.GetString("nname"), viper.GetInt("nidx"))
	en := NewConfig()
	en.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		en.Stop()
	})
	return nil
}
