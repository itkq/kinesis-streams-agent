package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/comail/colog"
	"github.com/itkq/kinesis-agent-go/aggregator"
	"github.com/itkq/kinesis-agent-go/api"
	"github.com/itkq/kinesis-agent-go/config"
	"github.com/itkq/kinesis-agent-go/file_watcher"
	"github.com/itkq/kinesis-agent-go/sender"
	"github.com/itkq/kinesis-agent-go/state"
	"github.com/itkq/kinesis-agent-go/version"
)

var (
	TrapSignals = []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}

	// Command line options
	configFile  string
	showVersion bool
)

func StartCLI() int {
	LogConfig()

	flag.StringVar(&configFile, "c", "", "configuration file (yaml) path")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Printf("kinesis-firehose-agent-go v%s\n", version.Version)
		return 0
	}

	if configFile == "" {
		log.Println("error: -c option (config file path) must be set.")
		return 1
	}

	// TODO: debug option

	var err error

	conf, err := config.LoadConfig(configFile)
	if err != nil {
		log.Println("error:", err)
		return 1
	}

	state, err := state.LoadFromJSON(conf.StateConfig.StateFilePath)
	if err != nil {
		log.Println("error:", err)
		return 1
	}

	aggregator := aggregator.NewAggregator(conf.AggregatorConfig)

	watcher, err := filewatcher.NewFileWatcher(
		conf.FileWatcherConfig,
		state,
		aggregator.ChunkCh,
	)
	if err != nil {
		log.Println("error:", err)
		return 1
	}

	sender, err := sender.NewSender(
		conf.SenderConfig,
		state,
		aggregator.PayloadCh,
	)
	if err != nil {
		log.Println("error:", err)
		return 1
	}

	api, err := api.NewAPI(conf.APIConfig.Address)
	if err != nil {
		log.Println("error:", err)
		return 1
	}

	api.Register(sender)

	controlCh := make(chan interface{})

	go api.Run()
	go aggregator.Run()
	go sender.Run()
	go watcher.Run(controlCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, TrapSignals...)

	// waiting signal
	sig := <-sigCh
	log.Printf("info: received signal (%s)\n", sig)

	// shutdown
	close(controlCh)

	return 0
}

func LogConfig() {
	colog.SetDefaultLevel(colog.LDebug)
	colog.SetMinLevel(colog.LTrace)
	colog.SetFormatter(&colog.StdFormatter{
		Colors: true,
		Flag:   log.Ldate | log.Ltime | log.Lshortfile,
	})
	colog.Register()
}
