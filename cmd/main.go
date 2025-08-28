package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	todo "github.com/balamuteon/todo_restapi"
	"github.com/balamuteon/todo_restapi/pkg/cache"
	"github.com/balamuteon/todo_restapi/pkg/handler"
	"github.com/balamuteon/todo_restapi/pkg/repository"
	"github.com/balamuteon/todo_restapi/pkg/service"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	// Инициализируем Viper для чтения переменных окружения
	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %s", err.Error())
	}

	// Добавляем цикл повторных попыток подключения к БД
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
			Password: viper.GetString("db.password"), // Viper теперь управляет и паролем
		})

		if err == nil {
			logrus.Info("Successfully connected to the database.")
			break // Успешно подключились, выходим из цикла
		}

		logrus.Warnf("Failed to connect to db, retrying in %v... (%d/%d)", retryDelay, i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	if err != nil {
		logrus.Fatalf("failed to initialize db after %d retries: %s", maxRetries, err.Error())
	}

	client, err := cache.NewRedisClient(&cache.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       viper.GetInt("redis.db"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize cache client: %s", err.Error())
	}

	cache := cache.NewCache(client)
	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services, cache)

	srv := new(todo.Server)

	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("TodoApp Started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("TodoApp Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := client.Close(); err != nil {
		logrus.Errorf("error occured on redis connection close: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		logrus.Warn("config file not found, relying on environment variables")
	}

	viper.AutomaticEnv()
	// Заменяем точки на подчеркивания для переменных (db.host -> DB_HOST)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("db.host", "db")
	viper.SetDefault("db.port", "5432")
	viper.SetDefault("db.sslmode", "disable")

	return nil
}
