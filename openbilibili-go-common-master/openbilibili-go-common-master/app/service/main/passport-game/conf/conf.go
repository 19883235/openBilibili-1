package conf

import (
	"errors"
	"flag"

	"go-common/app/service/main/passport-game/model"
	"go-common/library/cache/memcache"
	"go-common/library/conf"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/trace"
	"go-common/library/time"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	// Conf conf.
	Conf   = &Config{}
	client *conf.Client
)

// Config config.
type Config struct {
	// Proxy if proxy
	Proxy bool
	// URIs
	AccountURI  string
	PassportURI string
	// Xlog log
	Xlog *log.Config
	// Tracer tracer
	Tracer *trace.Config
	// DB db
	DB *DB
	// Memcache memcache
	Memcache *Memcache
	// HTTPClient http client
	HTTPClient *bm.ClientConfig
	// HTTP server config
	BM *bm.ServerConfig
	// Dispatcher dispatcher
	Dispatcher *Dispatcher
}

// Dispatcher dispatcher.
type Dispatcher struct {
	Name        string
	Oauth       map[string]string
	RenewToken  map[string]string
	RegionInfos []*model.RegionInfo
}

// DB db config.
type DB struct {
	Cloud       *sql.Config
	OtherRegion *sql.Config
}

// Memcache general memcache config.
type Memcache struct {
	*memcache.Config
	Expire time.Duration
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init config.
func Init() (err error) {
	if confPath != "" {
		return local()
	}
	return remote()
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	go func() {
		for range client.Event() {
			log.Info("config reload")
			if load() != nil {
				log.Error("config reload error (%v)", err)
			}
		}
	}()
	return
}

func load() (err error) {
	var (
		s       string
		ok      bool
		tmpConf *Config
	)
	if s, ok = client.Toml2(); !ok {
		return errors.New("load config center error")
	}

	if _, err = toml.Decode(s, &tmpConf); err != nil {
		return errors.New("could not decode config")
	}
	*Conf = *tmpConf
	return
}
