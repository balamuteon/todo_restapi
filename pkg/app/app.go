package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	todo "github.com/balamuteon/todo_restapi"
	"github.com/balamuteon/todo_restapi/pkg/cache"
	"github.com/balamuteon/todo_restapi/pkg/handler"
	"github.com/balamuteon/todo_restapi/pkg/repository"
	"github.com/balamuteon/todo_restapi/pkg/service"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type App struct {
	db       *sqlx.DB
	redis    *redis.Client
	services *service.Service
	cache    cache.Cache
}

// NewApp создает и инициализирует новый экземпляр приложения.
func NewApp() (*App, error) {
	db, err := connectToDBWithRetry()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %w", err)
	}

	client, err := cache.NewRedisClient(&cache.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize cache client: %w", err)
	}

	appCache := cache.NewCache(client)
	repos := repository.NewRepository(db)
	services := service.NewService(repos)

	return &App{
		db:       db,
		redis:    client,
		services: services,
		cache:    appCache,
	}, nil
}

// Run запускает HTTP-сервер и управляет его жизненным циклом.
func (a *App) Run() error {
	handlers := handler.NewHandler(a.services, a.cache)

	srv := new(todo.Server)
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("error occurred while running http server: %s", err.Error())
		}
	}()

	logrus.Print("TodoApp Started on port: ", viper.GetString("port"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("TodoApp Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("error occurred on server shutting down: %w", err)
	}

	if err := a.redis.Close(); err != nil {
		return fmt.Errorf("error occurred on redis connection close: %w", err)
	}

	if err := a.db.Close(); err != nil {
		return fmt.Errorf("error occurred on db connection close: %w", err)
	}

	return nil
}

// connectToDBWithRetry инкапсулирует логику подключения к БД с повторными попытками.
func connectToDBWithRetry() (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error
	maxRetries := 5
	retryDelay := time.Second * 5

	for i := 0; i < maxRetries; i++ {
		db, err = repository.NewPostgresDB(repository.Config{
			Host:     viper.GetString("db.host"),
			Port:     viper.GetString("db.port"),
			Username: viper.GetString("db.username"),
			DBName:   viper.GetString("db.dbname"),
			SSLMode:  viper.GetString("db.sslmode"),
			Password: viper.GetString("db.password"),
		})

		if err == nil {
			logrus.Info("Successfully connected to the database.")
			return db, nil
		}

		logrus.Warnf("Failed to connect to db, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	return nil, err
}
