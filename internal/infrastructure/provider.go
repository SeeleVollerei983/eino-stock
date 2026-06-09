package infrastructure

import (
	"os"
	"time"

	"eino-stock/internal/conf"
	"eino-stock/internal/infrastructure/cron"
	"eino-stock/internal/infrastructure/search"
	"eino-stock/internal/infrastructure/eastmoney"
	"eino-stock/internal/infrastructure/f10"
	"eino-stock/internal/infrastructure/quote"
	"eino-stock/internal/infrastructure/sina"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gopkg.in/yaml.v3"
)

var ProviderSet = wire.NewSet(
	NewQuoteClient,
	NewEastMoneyKLineClient,
	NewSinaKLineClient,
	NewScreenClient,
	NewSearchClient,
	NewF10Client,
	NewCronScheduler,
	NewSearchClient,
)

func NewSearchClient(c *conf.DataSource) *eastmoney.SearchClient {
	qgqpBId := readQgqpBId()
	if qgqpBId == "" {
		qgqpBId = os.Getenv("QGQP_B_ID")
	}
	return eastmoney.NewSearchClient(qgqpBId)
}

func NewQuoteClient(c *conf.DataSource, logger log.Logger) *quote.Client {
	timeout := 10 * time.Second
	if c != nil && c.HttpTimeout != nil {
		timeout = c.HttpTimeout.AsDuration()
	}
	return quote.NewClient(timeout, logger)
}

func NewEastMoneyKLineClient(c *conf.DataSource, logger log.Logger) *eastmoney.KLineClient {
	timeout := 10 * time.Second
	if c != nil && c.HttpTimeout != nil {
		timeout = c.HttpTimeout.AsDuration()
	}
	log.NewHelper(logger).Info("created eastmoney kline client")
	return eastmoney.NewKLineClient(timeout)
}

func NewSinaKLineClient(c *conf.DataSource, logger log.Logger) *sina.KLineClient {
	timeout := 10 * time.Second
	if c != nil && c.HttpTimeout != nil {
		timeout = c.HttpTimeout.AsDuration()
	}
	log.NewHelper(logger).Info("created sina kline client")
	return sina.NewKLineClient(timeout)
}

func NewScreenClient(c *conf.DataSource, logger log.Logger) *eastmoney.ScreenClient {
	timeout := 15 * time.Second
	if c != nil && c.HttpTimeout != nil {
		timeout = c.HttpTimeout.AsDuration()
	}
	qgqpBId := readQgqpBId()
	if qgqpBId == "" {
		qgqpBId = os.Getenv("QGQP_B_ID")
	}
	return eastmoney.NewScreenClient(qgqpBId, timeout)
}

func NewF10Client(c *conf.DataSource, logger log.Logger) *f10.Client {
	timeout := 10 * time.Second
	if c != nil && c.HttpTimeout != nil {
		timeout = c.HttpTimeout.AsDuration()
	}
	log.NewHelper(logger).Info("created f10 client")
	return f10.NewClient(timeout)
}

func NewDuckDuckGoSearchClient() *search.Client { return search.NewClient() }

func NewCronScheduler(logger log.Logger) *cron.Scheduler {
	log.NewHelper(logger).Info("created cron scheduler")
	return cron.NewScheduler()
}

func readQgqpBId() string {
	paths := []string{"configs/config.yaml", "../../configs/config.yaml", "../configs/config.yaml"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg struct {
			DataSource struct {
				QgqpBId string `yaml:"qgqp_b_id"`
			} `yaml:"data_source"`
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue
		}
		if cfg.DataSource.QgqpBId != "" {
			return cfg.DataSource.QgqpBId
		}
	}
	return ""
}




