package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	// Keep the America/Fortaleza timezone available in the scratch-like runtime image.
	_ "time/tzdata"

	"github.com/brocados-lanches/api/internal/config"
	"github.com/brocados-lanches/api/internal/httpapi"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
	"github.com/brocados-lanches/api/internal/repository"
	"github.com/brocados-lanches/api/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	startupContext, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()

	mongoClient, err := repository.Connect(startupContext, cfg.MongoURI)
	if err != nil {
		return err
	}
	defer func() {
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := mongoClient.Close(shutdownContext); err != nil {
			logger.Error("mongo disconnect failed", "error", err)
		}
	}()

	database := mongoClient.Database(cfg.MongoDB)
	if database == nil {
		return fmt.Errorf("mongo database %q is unavailable", cfg.MongoDB)
	}
	if err := repository.EnsureCatalogIndexes(startupContext, database); err != nil {
		return err
	}
	if err := repository.EnsureOrderIndexes(startupContext, database); err != nil {
		return err
	}
	if err := repository.EnsureCustomerProfileIndexes(startupContext, database); err != nil {
		return err
	}
	if err := repository.EnsureAdminIndexes(startupContext, database); err != nil {
		return err
	}
	if err := repository.EnsureAuditIndexes(startupContext, database); err != nil {
		return err
	}
	logger.Info("database indexes ensured", "database", cfg.MongoDB)

	catalogRepository, err := repository.NewCatalogRepository(database)
	if err != nil {
		return err
	}
	catalogService, err := service.NewCatalogService(catalogRepository)
	if err != nil {
		return err
	}
	adminCatalogService, err := service.NewAdminCatalogService(catalogRepository)
	if err != nil {
		return err
	}
	adminStoreService, err := service.NewAdminStoreService(catalogRepository)
	if err != nil {
		return err
	}
	operationLocation, err := time.LoadLocation("America/Fortaleza")
	if err != nil {
		return err
	}
	storeService, err := service.NewStoreService(catalogRepository, operationLocation)
	if err != nil {
		return err
	}

	orderRepository, err := repository.NewOrderRepository(database)
	if err != nil {
		return err
	}
	customerProfileRepository, err := repository.NewCustomerProfileRepository(database)
	if err != nil {
		return err
	}
	orderService, err := service.NewOrderService(
		orderRepository,
		operationLocation,
		service.WithCustomerProfiles(customerProfileRepository, logger),
	)
	if err != nil {
		return err
	}
	adminOrderService, err := service.NewAdminOrderService(
		orderRepository,
		service.WithAdminCustomerProfiles(customerProfileRepository),
	)
	if err != nil {
		return err
	}

	adminRepository, err := repository.NewAdminRepository(database)
	if err != nil {
		return err
	}
	adminAuthService, err := service.NewAdminAuthService(
		adminRepository,
		service.AdminSessionIdleDuration,
		service.AdminSessionAbsoluteDuration,
	)
	if err != nil {
		return err
	}
	adminUsersService, err := service.NewAdminUsersService(adminRepository)
	if err != nil {
		return err
	}
	if cfg.HasAdminBootstrap() {
		created, err := adminAuthService.BootstrapOwner(
			startupContext,
			cfg.AdminBootstrapName,
			cfg.AdminBootstrapEmail,
			cfg.AdminBootstrapPassword,
		)
		if err != nil {
			return err
		}
		if created {
			logger.Info("bootstrap admin owner created")
		}
	}
	loginLimiter, err := ratelimit.New(5, time.Minute)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr: cfg.Addr(),
		Handler: httpapi.NewRouter(httpapi.Dependencies{
			Pinger:       mongoClient,
			Catalog:      catalogService,
			Store:        storeService,
			Orders:       orderService,
			AdminAuth:    adminAuthService,
			AdminCatalog: adminCatalogService,
			AdminStore:   adminStoreService,
			AdminOrders:  adminOrderService,
			AdminUsers:   adminUsersService,
			AdminCookie: httpapi.AdminCookieConfig{
				Name:   cfg.AdminCookieNameEffective(),
				Secure: cfg.AdminCookieSecure,
			},
			LoginLimiter: loginLimiter,
			Limiters:     httpapi.DefaultLimiters(),
			Logger:       logger,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("api listening", "addr", cfg.Addr(), "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return err
	case signalReceived := <-stop:
		logger.Info("shutdown requested", "signal", signalReceived.String())
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	return server.Shutdown(shutdownContext)
}
