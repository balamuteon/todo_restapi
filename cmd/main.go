package main

import (
	"log"

	"github.com/balamuteon/todo_restapi"
	"github.com/balamuteon/todo_restapi/pkg/handler"
	"github.com/balamuteon/todo_restapi/pkg/repository"
	"github.com/balamuteon/todo_restapi/pkg/service"
)

func main() {
	repos := repository.NewRepository()
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(todo.Server)
	if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occured while running http server: %s", err.Error())
	}
}