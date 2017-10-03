package cli

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/comail/colog"
	"github.com/itkq/kinesis-agent-go/aggregator"
	"github.com/itkq/kinesis-agent-go/api"
	"github.com/itkq/kinesis-agent-go/config"
	"github.com/itkq/kinesis-agent-go/file_watcher"
	"github.com/itkq/kinesis-agent-go/sender"
	"github.com/itkq/kinesis-agent-go/sender/kinesis"
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

	awsConfig := aws.NewConfig()

	// configure forward proxy
	if conf.SenderConfig.ForwardProxyUrl != "" {
		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: func(*http.Request) (*url.URL, error) {
					return url.Parse(conf.SenderConfig.ForwardProxyUrl)
				},
			},
		}
		awsConfig = awsConfig.WithHTTPClient(httpClient)
		log.Println("info: configured forward proxy: ", conf.SenderConfig.ForwardProxyUrl)
	}

	ks, err := kinesis.NewKinesisStream(awsConfig)
	if err != nil {
		log.Println("error:", err)
		os.Exit(1)
	}
	sendClient := kinesis.NewKinesisStreamClient(ks, &conf.SenderConfig.StreamName)
	sender := sender.NewSender(sendClient, state, aggregator.PayloadCh)

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
