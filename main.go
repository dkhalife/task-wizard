package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"donetick.com/core/config"
	"donetick.com/core/frontend"
	auth "donetick.com/core/internal/middleware/auth"
	"donetick.com/core/internal/migrations"
	database "donetick.com/core/internal/utils/database"
	"donetick.com/core/internal/utils/email"
	utils "donetick.com/core/internal/utils/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	chore "donetick.com/core/internal/api/chore"
	"donetick.com/core/internal/api/label"
	user "donetick.com/core/internal/api/user"
	chRepo "donetick.com/core/internal/repos/chore"
	lRepo "donetick.com/core/internal/repos/label"
	nRepo "donetick.com/core/internal/repos/notifier"
	uRepo "donetick.com/core/internal/repos/user"
	logging "donetick.com/core/internal/services/logging"
	notifier "donetick.com/core/internal/services/notifications"
	"donetick.com/core/internal/services/planner"
	migration "donetick.com/core/internal/utils/migration"
)

func main() {
	logging.SetConfig(&logging.Config{
		Encoding:    "console",
		Level:       zapcore.Level(zapcore.DebugLevel),
		Development: true,
	})

	app := fx.New(
		fx.Supply(config.LoadConfig()),
		fx.Supply(logging.DefaultLogger().Desugar()),

		fx.Provide(auth.NewAuthMiddleware),

		fx.Provide(database.NewDatabase),
		fx.Provide(chRepo.NewChoreRepository),
		fx.Provide(chore.NewHandler),
		fx.Provide(uRepo.NewUserRepository),
		fx.Provide(user.NewHandler),

		fx.Provide(nRepo.NewNotificationRepository),
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
		fx.Provide(label.NewHandler),

		fx.Provide(frontend.NewHandler),

		fx.Invoke(
			chore.Routes,
			user.Routes,
			label.Routes,
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
	gin.SetMode(gin.DebugMode)
	// log when http request is made:

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
				err := migration.MigrationScripts(db, cfg)
				if err != nil {
					panic(err)
				}
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
