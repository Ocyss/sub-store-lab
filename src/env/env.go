package env

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var (
	Conf     envConfig
	DelayKey string
)

type envConfig struct {
	Host            string `env:"HOST" envDefault:"0.0.0.0"`
	Port            int    `env:"PORT" envDefault:"8000"`
	Debug           bool   `env:"DEBUG" envDefault:"false"`
	NoBeautifyNodes bool   `env:"NO_BEAUTIFY_NODES" envDefault:"false"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	DataDir         string `env:"DATA_DIR" envDefault:"./data"`
	Prefix          string `env:"PREFIX" envDefault:"_lab"`
	DelayTestUrl    string `env:"DELAY_TEST_URL" envDefault:"https://www.gstatic.com/generate_204"`
	EnableTester    string `env:"ENABLE_TESTER"`
	EnableEnvProxy  bool   `env:"ENABLE_ENV_PROXY" envDefault:"false"`

	IpQualityAPIKey  string `env:"IPQUALITY_API_KEY"`  // https://www.ipqualityscore.com/create-account
	AbuseIPDBAPIKey  string `env:"ABUSEIPDB_API_KEY"`  // https://www.abuseipdb.com/account
	IpregistryAPIKey string `env:"IPREGISTRY_API_KEY"` // https://dashboard.ipregistry.co/apikeys
	IpDataAPIKey     string `env:"IPDATA_API_KEY"`     // https://dashboard.ipdata.co/api.html
}

func init() {
	var err error
	if pwd, ok := os.LookupEnv("LAB_PWD"); ok {
		os.Chdir(pwd)
	}
	_ = godotenv.Load()
	// if err != nil {
	// 	dir, err := os.Getwd()
	// 	if err != nil {
	// 		slog.Warn("failed to get workdir", "error", err)
	// 	}
	// 	slog.Warn("failed to load env", "error", err, "workdir", dir)
	// }
	err = env.ParseWithOptions(&Conf, env.Options{
		Prefix:                "LAB_",
		UseFieldNameByDefault: true,
	})
	if err != nil {
		log.Fatalf("failed to parse env: %v", err)
	}
	Conf.DataDir, err = filepath.Abs(Conf.DataDir)
	if err != nil {
		log.Fatalf("failed to get absolute path: %v", err)
	}
	DelayKey = Conf.SnakeKey("delay")
	initLog()

	if Conf.Debug {
		slog.Debug("Conf", "Conf", Conf)
	}
}

func InitService() {
	os.MkdirAll(Conf.DataDir, 0o755)
	initDB()
}

func (e *envConfig) SnakeKey(args ...string) string {
	// snakeArgs := make([]string, len(args))
	// for i, arg := range args {
	// 	snakeArgs[i] = lo.SnakeCase(arg)
	// }
	return e.Prefix + "_" + strings.Join(args, "_")
}
