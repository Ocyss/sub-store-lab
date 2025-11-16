package env

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/ocyss/sub-store-lab/src/static"
	"gopkg.in/yaml.v3"
)

var (
	Conf         envConfig
	OverrideYaml map[string]any
)

// DelayKey string

type envConfig struct {
	Host string `env:"HOST" envDefault:"0.0.0.0"`
	Port int    `env:"PORT" envDefault:"8000"`

	Debug    bool   `env:"DEBUG" envDefault:"false"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	DataDir         string `env:"DATA_DIR" envDefault:"./data"`
	OutputNodesJson bool   `env:"OUTPUT_NODES_JSON" envDefault:"true"`

	// mihomoDNS配置路径，true使用内置dns `src/clash.yml`, false 禁用mihomoDNS
	EnableMihomoDNS string `env:"ENABLE_MIHOMO_DNS" envDefault:"true"`

	DelayTestUrl string `env:"DELAY_TEST_URL" envDefault:"https://www.gstatic.com/generate_204"`

	DisableTester string `env:"DISABLE_TESTER"` // 逗号分割, 不区分大小写，默认不禁用: Purity,Speed

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
	// DelayKey = Conf.SnakeKey("delay")
	initLog()

	if Conf.Debug {
		slog.Debug("Conf", "Conf", Conf)
	}
}

func InitService() {
	if err := os.MkdirAll(Conf.DataDir, 0o755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	overridePath := filepath.Join(Conf.DataDir, "override.yaml")
	f, err := os.OpenFile(overridePath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Fatalf("failed to open override.yaml: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		log.Fatalf("failed to stat override.yaml: %v", err)
	}

	if stat.Size() == 0 {
		// 文件为空则写入默认内容
		if _, err := f.Write(static.OverrideYamlByte); err != nil {
			log.Fatalf("failed to write default override.yaml: %v", err)
		}
		// 写入后重置读指针
		if _, err := f.Seek(0, 0); err != nil {
			log.Fatalf("failed to seek override.yaml: %v", err)
		}
	}

	overrideYaml, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("failed to read override.yaml: %v", err)
	}

	err = yaml.Unmarshal(overrideYaml, &OverrideYaml)
	if err != nil {
		log.Fatalf("failed to unmarshal override.yaml: %v", err)
	}

	initDB()
}

// func (e *envConfig) SnakeKey(args ...string) string {
// 	// snakeArgs := make([]string, len(args))
// 	// for i, arg := range args {
// 	// 	snakeArgs[i] = lo.SnakeCase(arg)
// 	// }
// 	return e.Prefix + "_" + strings.Join(args, "_")
// }
