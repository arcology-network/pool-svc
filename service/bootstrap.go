package service

import (
	"net/http"

	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/kafka"
	"github.com/arcology-network/component-lib/storage"
	"github.com/arcology-network/component-lib/streamer"
	"github.com/arcology-network/pool-svc/service/workers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type Config struct {
	concurrency int
	groupid     string
}

//return a Subscriber struct
func NewConfig() *Config {
	return &Config{
		concurrency: viper.GetInt("concurrency"),
		groupid:     "pool",
	}
}

func (cfg *Config) Start() {
	http.Handle("/streamer", promhttp.Handler())
	go http.ListenAndServe(":19012", nil)

	broker := streamer.NewStatefulStreamer()
	//00 initializer
	initializer := actor.NewActor(
		"initializer",
		broker,
		[]string{actor.MsgStarting},
		[]string{
			actor.MsgStartSub,
		},
		[]int{1},
		workers.NewInitializer(cfg.concurrency, cfg.groupid),
	)
	initializer.Connect(streamer.NewDisjunctions(initializer, 1))

	receiveMsgs := []string{
		actor.MsgReapinglist,
		actor.MsgInitReapCommand,
		actor.MsgReapCommand,
	}

	receiveTopics := []string{
		viper.GetString("reaping-list"),
		viper.GetString("msgexch"),
	}

	//01 kafkaDownloader
	kafkaDownloader := actor.NewActor(
		"kafkaDownloader",
		broker,
		[]string{actor.MsgStartSub},
		receiveMsgs,
		[]int{1, 1, 1},
		kafka.NewKafkaDownloader(cfg.concurrency, cfg.groupid, receiveTopics, receiveMsgs, viper.GetString("mqaddr")),
	)
	kafkaDownloader.Connect(streamer.NewDisjunctions(kafkaDownloader, 3))

	//01 kafkaDownloader
	kafkaDownloaderMsg := actor.NewActor(
		"kafkaDownloaderMsg",
		broker,
		[]string{actor.MsgStartSub},
		[]string{actor.MsgMessager},
		[]int{10000},
		kafka.NewKafkaDownloader(cfg.concurrency, cfg.groupid, []string{viper.GetString("chkd-message")}, []string{actor.MsgMessager}, viper.GetString("mqaddr2")),
	)
	kafkaDownloaderMsg.Connect(streamer.NewDisjunctions(kafkaDownloaderMsg, 10000))

	//02-00 Combiner
	combiner := actor.NewActor(
		"combiner",
		broker,
		[]string{
			actor.MsgReapCommand,
			actor.MsgClearCompleted,
		},
		[]string{
			actor.MsgStartReapCommand,
		},
		[]int{1},
		storage.NewCombiner(cfg.concurrency, cfg.groupid, actor.MsgStartReapCommand),
	)
	combiner.Connect(streamer.NewConjunctions(combiner))

	//02 aggreSelector
	aggreSelector := actor.NewActor(
		"aggreSelector",
		broker,
		[]string{
			actor.MsgMessager,
			actor.MsgClearCommand,
			actor.MsgReapinglist,
			actor.MsgInitReapCommand,
			actor.MsgStartReapCommand,
		},
		[]string{
			actor.MsgMessagersReaped,
			actor.MsgMetaBlock,
			actor.MsgSelectedTx,
			actor.MsgClearCompleted,
			actor.MsgClearCommand,
		},
		[]int{1, 1, 1, 1, 1},
		workers.NewAggreSelector(cfg.concurrency, cfg.groupid, viper.GetInt("rpm")),
	)
	aggreSelector.Connect(streamer.NewDisjunctions(aggreSelector, 1))
	//uploader
	relations := map[string]string{}
	relations[actor.MsgMessagersReaped] = viper.GetString("selected-msgs")
	relations[actor.MsgSelectedTx] = viper.GetString("selected-txs")
	relations[actor.MsgMetaBlock] = viper.GetString("meta-block")
	kafkaUploader := actor.NewActor(
		"kafkaUploader",
		broker,
		[]string{
			actor.MsgMessagersReaped,
			actor.MsgMetaBlock,
			actor.MsgSelectedTx,
		},
		[]string{},
		[]int{},
		kafka.NewKafkaUploader(cfg.concurrency, cfg.groupid, relations, viper.GetString("mqaddr")),
	)
	kafkaUploader.Connect(streamer.NewDisjunctions(kafkaUploader, 1))

	selfStarter := streamer.NewDefaultProducer("selfStarter", []string{actor.MsgStarting}, []int{1})
	broker.RegisterProducer(selfStarter)
	broker.Serve()
	//start signal
	streamerStarting := actor.Message{
		Name:   actor.MsgStarting,
		Height: 0,
		Round:  0,
		Data:   "start",
	}
	broker.Send(actor.MsgStarting, &streamerStarting)

}

func (cfg *Config) Stop() {

}
