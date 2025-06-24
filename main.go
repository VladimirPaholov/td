package main

import (
	"log"

	"github.com/vladimirpaholov/td/pkg/api"
	"github.com/vladimirpaholov/td/pkg/db"
	"github.com/vladimirpaholov/td/pkg/server"
)

func main() {
	// Сначала инициализируем базу
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}

	// Затем регистрируем API-обработчики
	api.Init()

	// И запускаем сервер
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

/*
go test -run ^TestApp$ ./tests
go test -run ^TestDB$ ./tests
go test -run ^TestNextDate$ ./tests
go test -run ^TestAddTask$ ./tests
go test -run ^TestTasks$ ./tests
go test -run ^TestEditTask$ ./tests
go test -run ^TestDone$ ./tests
*/
