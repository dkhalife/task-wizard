package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"dkhalife.com/tasks/core/backend"
	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/frontend"
	auth "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/migrations"
	database "dkhalife.com/tasks/core/internal/utils/database"
	"dkhalife.com/tasks/core/internal/utils/email"
	utils "dkhalife.com/tasks/core/internal/utils/middleware"
	ws "dkhalife.com/tasks/core/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"

	apis "dkhalife.com/tasks/core/internal/apis"
	cRepo "dkhalife.com/tasks/core/internal/repos/caldav"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/housekeeper"
	lService "dkhalife.com/tasks/core/internal/services/labels"
	logging "dkhalife.com/tasks/core/internal/services/logging"
	notifier "dkhalife.com/tasks/core/internal/services/notifications"
	"dkhalife.com/tasks/core/internal/services/scheduler"
	tService "dkhalife.com/tasks/core/internal/services/tasks"
	uService "dkhalife.com/tasks/core/internal/services/users"
	migration "dkhalife.com/tasks/core/internal/utils/migration"
)

func main() {
	cfgFile := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg := config.LoadConfig(*cfgFile)
	level, err := zapcore.ParseLevel(cfg.Server.LogLevel)
	if err != nil {
		level = zapcore.WarnLevel
	}

	logging.SetConfig(&logging.Config{
		Encoding:    "console",
		Level:       level,
		Development: level == zapcore.DebugLevel,
	})

	app := fx.New(
		fx.Supply(cfg),
		fx.Supply(logging.DefaultLogger().Desugar()),
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.NopLogger
		}),

		fx.Provide(auth.NewAuthMiddleware),

		fx.Provide(database.NewDatabase),
		fx.Provide(tRepo.NewTaskRepository),
		fx.Provide(apis.TasksAPI),
		fx.Provide(uRepo.NewUserRepository),
		fx.Provide(nRepo.NewNotificationRepository),
		fx.Provide(apis.UsersAPI),

		// add services
		fx.Provide(notifier.NewNotifier),
		fx.Provide(housekeeper.NewPasswordResetCleaner),
		fx.Provide(housekeeper.NewAppTokenCleaner),

		// Rate limiter
		fx.Provide(utils.NewRateLimiter),

		// add email sender:
		fx.Provide(email.NewEmailSender),
		// add handlers also
		fx.Provide(newServer),
		fx.Provide(ws.NewWSServer),
		fx.Provide(scheduler.NewScheduler),

		// Labels:
		fx.Provide(cRepo.NewCalDavRepository),
		fx.Provide(lRepo.NewLabelRepository),
		fx.Provide(lService.NewLabelService),
		fx.Provide(lService.NewLabelsMessageHandler),
		fx.Provide(uService.NewUserService),
		fx.Provide(uService.NewUsersMessageHandler),
		fx.Provide(tService.NewTaskService),
		fx.Provide(tService.NewTasksMessageHandler),
		fx.Provide(apis.LabelsAPI),
		fx.Provide(apis.LogsAPI),
		fx.Provide(apis.CalDAVAPI),

		fx.Provide(frontend.NewHandler),
		fx.Provide(backend.NewHandler),

		fx.Invoke(
			apis.TaskRoutes,
			apis.UserRoutes,
			apis.LabelRoutes,
			apis.CalDAVRoutes,
			apis.LogRoutes,
			ws.Routes,
			tService.TaskMessages,
			lService.LabelMessages,
			uService.UserMessages,
			frontend.Routes,
			backend.Routes,
		),
	)

	if err := app.Err(); err != nil {
		log.Fatal(err)
	}

	app.Run()

}

func newServer(lc fx.Lifecycle, cfg *config.Config, db *gorm.DB, bgScheduler *scheduler.Scheduler) *gin.Engine {
	if cfg.Server.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      h2c.NewHandler(r, h2s),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
		log.Fatalf("failed to configure HTTP/2 server: %v", err)
	}
	if len(cfg.Server.AllowedOrigins) > 0 {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowOrigins = cfg.Server.AllowedOrigins
		if cfg.Server.AllowCorsCredentials {
			corsCfg.AllowCredentials = true
		}
		corsCfg.AddAllowHeaders("Authorization")
		r.Use(cors.New(corsCfg))
	}
	r.Use(utils.RequestLogger())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logging.FromContext(ctx).Info("Starting server")

			if cfg.Database.Migration {
				if err := migration.Migration(db); err != nil {
					return fmt.Errorf("failed to auto-migrate: %s", err.Error())
				}

				if err := migrations.Run(ctx, db); err != nil {
					return fmt.Errorf("failed to run migrations: %s", err.Error())
				}
			}

			bgScheduler.Start(context.Background())

			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log := logging.FromContext(ctx)
					log.Fatalf("listen: %s\n", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			bgScheduler.Stop()

			if err := srv.Shutdown(ctx); err != nil {
				log := logging.FromContext(ctx)
				log.Fatalf("Server Shutdown: %s", err)
			} else {
				log := logging.FromContext(ctx)
				log.Info("Server stopped")
			}
			return nil
		},
	})

	return r
}
