package app

import (
	"context"
	"github.com/deissh/osu-lazer/ayako/api"
	"github.com/deissh/osu-lazer/ayako/config"
	"github.com/deissh/osu-lazer/ayako/middlewares/customlogger"
	"github.com/deissh/osu-lazer/ayako/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"time"
)

type App struct {
	Config *config.Config
	Store  store.Store
	Echo   *echo.Echo

	goroutineCount      int32
	goroutineExitSignal chan struct{}
}

// NewApp with DI
// expect store.Store
func NewApp(cfg *config.Config, store store.Store) *App {
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true

	e.Use(middleware.RequestID())
	e.Use(customlogger.Middleware())

	api.New(store, e.Group("/v2"))

	s := &App{
		Store:  store,
		Echo:   e,
		Config: cfg,
	}
	s.Config.AutoReloadCallback = s.OnConfigReload

	if cfg.Server.EnableJobs {
		s.Go(func() {
			runSecurityJob(s)
		})
		s.Go(func() {
			runUpdateCheck(s)
		})
	}

	return s
}

// Start http server and setup graceful shutdown
func (s *App) Start() error {
	for _, route := range s.Echo.Routes() {
		log.Debug().
			Str("method", route.Method).
			Str("path", route.Path).
			Msg("Route loaded")
	}

	log.Info().Msg("Starting App...")

	addr := s.Config.Server.Host + ":" + s.Config.Server.Port

	// Graceful start and stop HTTP server
	go func() {
		err := s.Echo.Start(addr)
		if err != nil {
			log.Error().
				Err(err).
				Msg("shutting down the server")
		}
	}()

	log.Info().Msg("App started on http://" + addr)

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		cancel()
		s.WaitForGoroutines()
	}()

	return s.Echo.Shutdown(ctx)
}

func runSecurityJob(s *App) {
	doSecurity(s)
	CreateRecurringTask("Security", func() {
		doSecurity(s)
	}, time.Hour*6)
}

func runUpdateCheck(s *App) {
	doUpdateCheck(s)
	CreateRecurringTask("UpdateCheck", func() {
		doUpdateCheck(s)
	}, time.Hour)
}

func doSecurity(s *App) {
	s.DoSecurityUpdateCheck()
}

func doUpdateCheck(s *App) {
	s.DoBeatmapSetUpdate()
}
