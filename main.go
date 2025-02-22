package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"dkhalife.com/tasks/core/config"
	"dkhalife.com/tasks/core/frontend"
	auth "dkhalife.com/tasks/core/internal/middleware/auth"
	"dkhalife.com/tasks/core/internal/migrations"
	database "dkhalife.com/tasks/core/internal/utils/database"
	"dkhalife.com/tasks/core/internal/utils/email"
	utils "dkhalife.com/tasks/core/internal/utils/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	apis "dkhalife.com/tasks/core/internal/apis"
	lRepo "dkhalife.com/tasks/core/internal/repos/label"
	nRepo "dkhalife.com/tasks/core/internal/repos/notifier"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	logging "dkhalife.com/tasks/core/internal/services/logging"
	notifier "dkhalife.com/tasks/core/internal/services/notifications"
	"dkhalife.com/tasks/core/internal/services/planner"
	migration "dkhalife.com/tasks/core/internal/utils/migration"
)

func main() {
	if os.Getenv("TW_ENV") == "debug" {
		logging.SetConfig(&logging.Config{
			Encoding:    "console",
			Level:       zapcore.Level(zapcore.DebugLevel),
			Development: true,
		})
	} else {
		logging.SetConfig(&logging.Config{
			Encoding:    "console",
			Level:       zapcore.Level(zapcore.WarnLevel),
			Development: false,
		})
	}

	app := fx.New(
		fx.Supply(config.LoadConfig()),
		fx.Supply(logging.DefaultLogger().Desugar()),

		fx.Provide(auth.NewAuthMiddleware),

		fx.Provide(database.NewDatabase),
		fx.Provide(tRepo.NewTaskRepository),
		fx.Provide(apis.TasksAPI),
		fx.Provide(uRepo.NewUserRepository),
		fx.Provide(nRepo.NewNotificationRepository),
		fx.Provide(apis.UsersAPI),

		fx.Provide(planner.NewNotificationPlanner),

		// add notifier
		fx.Provide(notifier.NewNotifier),

		// Rate limiter
		fx.Provide(utils.NewRateLimiter),

		// add email sender:
		fx.Provide(email.NewEmailSender),
		// add handlers also
		fx.Provide(newServer),
		fx.Provide(notifier.NewScheduler),

		// Labels:
		fx.Provide(lRepo.NewLabelRepository),
		fx.Provide(apis.LabelsAPI),

		fx.Provide(frontend.NewHandler),

		fx.Invoke(
			apis.TaskRoutes,
			apis.UserRoutes,
			apis.LabelRoutes,
			frontend.Routes,

			func(r *gin.Engine) {},
		),
	)

	if err := app.Err(); err != nil {
		log.Fatal(err)
	}

	app.Run()

}

func newServer(lc fx.Lifecycle, cfg *config.Config, db *gorm.DB, notifier *notifier.Scheduler) *gin.Engine {
	if os.Getenv("TW_ENV") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AddAllowHeaders("Authorization", "secretkey")
	r.Use(cors.New(config))

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if cfg.Database.Migration {
				migration.Migration(db)
				migrations.Run(context.Background(), db)
			}
			notifier.Start(context.Background())
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("listen: %s\n", err)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Fatalf("Server Shutdown: %s", err)
			}
			return nil
		},
	})

	return r
}
