package main

import (
	"context"
	"github.com/HJyup/mlt-gateway/internal/gateway"
	"github.com/HJyup/mlt-gateway/internal/handler"
	"github.com/HJyup/mtl-common"
	"github.com/HJyup/mtl-common/consul"
	mux2 "github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
)

type Specification struct {
	ServiceName string `required:"true" default:"gateway"`
	Address     string `required:"true"`
	Consul      string `required:"true"`
	Environment string `required:"true"`
}

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer func() {
		_ = logger.Sync()
	}()

	var s Specification
	if err = envconfig.Process("gateway", &s); err != nil {
		logger.Fatal("Failed to process environment variables", zap.Error(err))
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	registry, err := consul.NewRegistry(s.Consul)
	if err != nil {
		logger.Fatal("Failed to create registry", zap.Error(err))
	}

	instanceID := common.GenerateInstanceID(s.ServiceName)
	if err = registry.Register(instanceID, s.ServiceName, s.Address); err != nil {
		logger.Fatal("Failed to register service", zap.Error(err))
	}
	defer func(registry *consul.Registry, instanceID string) {
		err = registry.DeRegister(instanceID)
		if err != nil {
			logger.Fatal("Failed to deregister service: %v", zap.Error(err))
		}
	}(registry, instanceID)

	router := mux2.NewRouter()

	userGateway := gateway.NewUserGateway(registry, logger)
	userHandler := handler.NewUserHandler(userGateway)
	userHandler.RegisterRoutes(router)

	configGateway := gateway.NewConfigurationGateway(registry, logger)
	configHandler := handler.NewConfigurationHandler(configGateway)
	configHandler.RegisterRoutes(router)

	agentGateway := gateway.NewAgentGateway(registry, logger)
	agentHandler := handler.NewAgentHandler(agentGateway)
	agentHandler.RegisterRoutes(router)

	logger.Info("Starting server", zap.String("address", s.Address))
	if err = http.ListenAndServe(s.Address, router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
