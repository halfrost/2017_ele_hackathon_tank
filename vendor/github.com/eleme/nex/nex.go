package nex

import (
	"flag"

	"github.com/damnever/cc"
	huskarpool "github.com/eleme/huskar-pool"
	"github.com/eleme/huskar/config"
	"github.com/eleme/huskar/service"
	"github.com/eleme/huskar/toggle"
	"github.com/eleme/nex/app"
	"github.com/eleme/nex/client/kafka"
	"github.com/eleme/nex/client/redis"
	"github.com/eleme/nex/consts"
	"github.com/eleme/nex/consts/huskarkeys"
	"github.com/eleme/nex/db"
	"github.com/eleme/nex/log"
	"github.com/eleme/nex/metric"
	"github.com/eleme/nex/mock"
	"github.com/eleme/nex/tracking/etrace"
	"github.com/eleme/nex/utils"
	"github.com/eleme/samaritan/sdk"

	"github.com/apache/thrift/lib/go/thrift"
)

var (
	dev         = flag.Bool("dev", false, "Used for development environment.")
	addr        = flag.String("addr", "", "The address that server serve on.")
	localHuskar = flag.Bool("local-huskar", false, "Read local huskar config(config.json) service(service.json) instead of reading from huskar api. Used for development without access to huskar")
)

const (
	localHuskarConfigFile  = "config.json"
	localHuskarToggleFile  = "toggle.json"
	localHuskarServiceFile = "service.json"

	huskarConfigCachePath  = "/data/run/nex_huskar/config.json"
	huskarToggleCachePath  = "/data/run/nex_huskar/toggle.json"
	huskarServiceCachePath = "/data/run/nex_huskar/service.json"
)

// Init initialize necessary resources.
func Init() cc.Configer {
	hub.Lock()
	defer hub.Unlock()
	if hub.Inited {
		panic(ErrAlreadyInited)
	}
	hub.Inited = true

	var err error
	// Parse command line arguments
	flag.Parse()

	// Init config
	hub.NexConfig, err = initNexConfig()
	utils.Must(err)
	plugins := hub.NexConfig.Config("plugins")
	appName := hub.NexConfig.String("app_name")
	hub.NexConfig.SetDefault("addr", "0.0.0.0:8010")

	{ // Init logger
		inContainer := hub.NexConfig.BoolOr(consts.EnvDockerContainerID, false)
		addr := ""
		if !*dev && !inContainer {
			addr = hub.NexConfig.StringOr("syslog_addr", ":514")
		}
		log.Setup(appName, !inContainer, addr)
	}

	logger, err := log.GetContextLogger("nex")
	utils.Must(err)

	{ // Init Huskar Config/Toggle/Register
		if *localHuskar {
			if !*dev {
				logger.Warn("Running without dev mode but with local huskar config enabled!! Don't run like this in prod!!")
			}
			hub.HuskarConfiger, err = mock.NewHuskarFileConfiger(localHuskarConfigFile)
			utils.Must(err, "init Huskar config failed")
			hub.HuskarToggler, err = mock.NewHuskarFileToggler(localHuskarToggleFile)
			utils.Must(err, "init Huskar toggler failed")
			hub.HuskarRegistrator = mock.NewHuskarFileReisgrator(localHuskarServiceFile)
		} else {
			huskarConfig := huskarConfigFromNex(hub.NexConfig)
			hub.HuskarConfiger, err = config.New(huskarConfig)
			utils.Must(err, "init Huskar config failed")
			utils.Must(hub.HuskarConfiger.EnableCache(huskarConfigCachePath), "enable Huskar config cache")
			hub.HuskarConfiger.SetLogger(logger)
			hub.HuskarToggler, err = toggle.New(huskarConfig)
			utils.Must(err, "init Huskar toggler failed")
			utils.Must(hub.HuskarToggler.EnableCache(huskarToggleCachePath), "enable Huskar toggle cache")
			hub.HuskarRegistrator, err = service.New(huskarConfig)
			utils.Must(err, "init Huskar service failed")
			utils.Must(hub.HuskarRegistrator.EnableCache(huskarServiceCachePath), "enable Huskar service cache")
			hub.HuskarRegistrator.SetLogger(logger)
		}
	}

	{ // Init Huskar Pool
		hub.HuskarPool = huskarpool.NewWithService(hub.HuskarRegistrator)
	}

	// Init statsd
	plugins.SetDefault("statsd", !(*dev))
	if plugins.Bool("statsd") {
		statsdAddr := hub.NexConfig.String("statsd_url")
		if statsdAddr != "" {
			utils.Must(metric.StartWithOptions(statsdAddr, false, false), "start statsd failed")
		}
	}

	// Init etrace
	plugins.SetDefault("etrace", !(*dev))
	if plugins.Bool("etrace") {
		hub.ETrace, err = etrace.New(hub.NexConfig)
		utils.Must(err, "init etrace failed")
	}

	// Init Reids
	if plugins.BoolOr("redis", true) {
		logger, err := log.GetContextLogger("redis")
		utils.Must(err)
		redisSettings, err := hub.HuskarConfiger.Get(huskarkeys.RedisSettings)
		utils.Must(err, "get", huskarkeys.RedisSettings, "from Huskar failed")
		// TODO: or get redis settings from huskar and merge into nexconfig.Configer
		pm, err := redis.NewPoolManager(logger, redisSettings)
		utils.Must(err, "init redis failed")
		hub.RedisPools = pm
	}

	// Init DB
	if plugins.BoolOr("db", true) {
		logger, err := log.GetContextLogger("db")
		utils.Must(err)
		dbSettings, err := hub.HuskarConfiger.Get(huskarkeys.DBSettings)
		utils.Must(err, "get", huskarkeys.DBSettings, "from Huskar failed")
		dbm, err := db.NewDBManager(appName, logger, dbSettings)
		utils.Must(err, "init db failed")
		hub.DBManager = dbm
	}

	if !*dev && plugins.BoolOr("sam", false) { // Init sam client
		hub.SamClient, err = sdk.NewLocalClient(appName, hub.NexConfig.String("cluster"))
		utils.Must(err, "Init Samaritan client fialed") // XXX: fatal??
		utils.Must(hub.SamClient.DeclareUserApplication())
	}

	if plugins.BoolOr("kafka", false) {
		logger, err := log.GetContextLogger("kafka")
		utils.Must(err)
		jsonKafkaSettings, err := hub.HuskarConfiger.Get(huskarkeys.KafkaSettings)
		utils.Must(err, "get", huskarkeys.KafkaSettings, "from Huskar failed")
		client, err := kafka.New(hub.NexConfig, jsonKafkaSettings, logger)
		utils.Must(err, "init kafka client failed")
		hub.KafkaClient = client
	}

	return hub.NexConfig
}

// Serve starts a Thrift Application with given processor.
func Serve(processorFactory thrift.TProcessorFactory) {
	logger, err := log.GetContextLogger("nex")
	utils.Must(err)

	thriftApp := app.NewThriftApplication(
		GetNexConfig(),
		logger,
		GetHuskarRegistrator(),
		processorFactory,
	)
	utils.Must(thriftApp.Run())
}
