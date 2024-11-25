package main

import (
	"context"
	"edge-app/api/middlewares"
	"edge-app/api/routers"
	"edge-app/configs"
	"edge-app/pkg/kafka/consumer"
	"edge-app/pkg/kafka/producer"
	"edge-app/pkg/logging"
	"edge-app/pkg/metrics"
	"edge-app/pkg/traces"
	"fmt"
	"os"
	"strconv"

	"github.com/dimiro1/banner"
	"github.com/mattn/go-colorable"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg := configs.Get()
	setUpBanner(cfg)

	cleanup := traces.InitTracer(cfg)
	defer func(ctx context.Context) {
		err := cleanup(ctx)
		if err != nil {
			fmt.Printf("Error cleaning up traces: %v\n", err)
		}
	}(context.Background())

	gin.SetMode(cfg.Server.RunMode)
	r := gin.New()

	r.Use(middlewares.DefaultLogger(cfg))
	r.Use(middlewares.Authentication(cfg))
	r.Use(middlewares.Authorization(cfg))
	r.Use(middlewares.Prometheus())
	r.Use(otelgin.Middleware(cfg.Application.Name))
	r.Use(gin.CustomRecovery(middlewares.ErrorHandler))

	registerPrometheus()
	registerRouts(r)

	p := producer.NewProducible(cfg)
	defer p.Close()

	c := consumer.NewConsumable(cfg)
	defer c.Close()

	err := r.Run(":" + strconv.Itoa(cfg.Port))
	if err != nil {
		panic(err)
	}
}

func setUpBanner(cfg *configs.Config) {
	file, err := os.Open(cfg.Banner.FilePath)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(file)

	banner.Init(colorable.NewColorableStdout(), true, true, file)
}

func registerRouts(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	api := r.Group("/api")
	routers.Health(api.Group("/v1"))
	routers.BaseRouter(api.Group("/v1"))
}

func registerPrometheus() {
	logger := logging.NewLogger(configs.Get())

	err := prometheus.Register(metrics.HttpCall)
	if err != nil {
		logger.Error(logging.Prometheus, logging.Startup, err.Error(), nil)
	}

	err = prometheus.Register(metrics.HttpDuration)
	if err != nil {
		logger.Error(logging.Prometheus, logging.Startup, err.Error(), nil)
	}
}
