package main

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/khiki1995/todo/config"
	"github.com/khiki1995/todo/routes"
)

func setupRoutes(app *fiber.App) {
	routes.RouteList(app.Group("/api/todo"))
	routes.RouteListItem(app.Group("/api/todo"))
}

func main() {
	app := fiber.New()
	config.ConnectDB()

	app.Use(cors.New())
	app.Use(logger.New())

	setupRoutes(app)
	err := app.Listen(":3000")
	if err != nil {
		panic(errors.New("Error app failed to start "))
	}
}

//list list:
//1. Название
//2. Описание
//3. Цвет (для отображения на фронтенде)
//
//list item:
//4. Формулировку
//5. [опционально] Дату
//6. [опционально] Приоритет
//
//GetTodoList
//GetTodoListItem
//Пользователь может сортировать элементы списка (новый порядок сохраняется)
//Пользователь может обновлять/удалять элементы списков/списки
//
// Реализовать при помощи фреймворка Gin/Fiber(Golang) и MongoDB с применением docker-compose, api задокументировать в postman/swagger
// Docker
// Тесты
// Документация swagger/postman