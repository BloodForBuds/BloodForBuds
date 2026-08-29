package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/BloodForBuds/BloodForBuds/services/server/internal/app_store"
	apphttp "github.com/BloodForBuds/BloodForBuds/services/server/internal/httpserver"
	"github.com/BloodForBuds/BloodForBuds/services/server/internal/key_store"
	"github.com/BloodForBuds/BloodForBuds/services/server/internal/kms"
)

const defaultAddress = ":8080"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	startupCtx, cancelStartup := context.WithTimeout(ctx, 30*time.Second)
	defer cancelStartup()

	appStoreConfig, err := appStoreConfigFromEnvironment()
	if err != nil {
		return err
	}
	appStore, err := app_store.Open(startupCtx, appStoreConfig)
	if err != nil {
		return err
	}
	defer appStore.Close()
	if err := appStore.Migrate(startupCtx); err != nil {
		return err
	}

	keyStoreConfig, err := keyStoreConfigFromEnvironment()
	if err != nil {
		return err
	}
	keyStore, err := key_store.Open(startupCtx, keyStoreConfig)
	if err != nil {
		return err
	}
	defer keyStore.Close()
	if err := keyStore.Migrate(startupCtx); err != nil {
		return err
	}

	kmsClient, err := kms.New(kms.Config{
		Address: os.Getenv("BAO_ADDR"),
		Token:   os.Getenv("BAO_TOKEN"),
	})
	if err != nil {
		return err
	}
	kmsHealth, err := kmsClient.Health(startupCtx)
	if err != nil {
		return err
	}
	log.Printf(
		"connected to OpenBao version=%s initialized=%t sealed=%t",
		kmsHealth.Version,
		kmsHealth.Initialized,
		kmsHealth.Sealed,
	)

	server := &http.Server{
		Addr:              defaultAddress,
		Handler:           apphttp.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		return nil
	}
}

func appStoreConfigFromEnvironment() (app_store.Config, error) {
	port, err := environmentPort("APP_DB_PORT")
	if err != nil {
		return app_store.Config{}, err
	}
	return app_store.Config{
		Host:     os.Getenv("APP_DB_HOST"),
		Port:     port,
		Database: os.Getenv("APP_DB_NAME"),
		User:     os.Getenv("APP_DB_USER"),
		Password: os.Getenv("APP_DB_PASSWORD"),
		SSLMode:  environmentOr("APP_DB_SSLMODE", "disable"),
	}, nil
}

func keyStoreConfigFromEnvironment() (key_store.Config, error) {
	port, err := environmentPort("KEY_DB_PORT")
	if err != nil {
		return key_store.Config{}, err
	}
	return key_store.Config{
		Host:     os.Getenv("KEY_DB_HOST"),
		Port:     port,
		Database: os.Getenv("KEY_DB_NAME"),
		User:     os.Getenv("KEY_DB_USER"),
		Password: os.Getenv("KEY_DB_PASSWORD"),
		SSLMode:  environmentOr("KEY_DB_SSLMODE", "disable"),
	}, nil
}

func environmentPort(name string) (int, error) {
	value := environmentOr(name, "5432")
	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return port, nil
}

func environmentOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
