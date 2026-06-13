package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	carshop "utfpr.edu.br/carshop-api"
	"utfpr.edu.br/carshop-api/src/config"
	"utfpr.edu.br/carshop-api/src/controller"
	"utfpr.edu.br/carshop-api/src/infra/db"
	"utfpr.edu.br/carshop-api/src/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("startup: %v", err)
	}
}

func run() error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	pool, err := db.Open(ctx, settings.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, carshop.MigrationsFS, "migrations"); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	jwt := service.NewJWTService(settings.JWTSecret, settings.JWTIssuer, settings.JWTAudience, settings.JWTExpiryMin)
	users := service.NewUserService(pool)
	cars := service.NewCarService(pool)
	orders := service.NewOrderService(pool)
	comm := service.NewComissionService(pool)

	if err := users.SeedAdmin(ctx, settings.SuperUserUsername, settings.SuperUserPassword, settings.SuperUserEmail); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	router := controller.BuildRouter(controller.Deps{
		JWT:       jwt,
		Users:     users,
		Cars:      cars,
		Orders:    orders,
		Comission: comm,
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", settings.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("listen: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Println("server stopped cleanly")
	return nil
}
