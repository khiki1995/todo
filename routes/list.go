package routes

import (
	"github.com/gofiber/fiber/v2"
	controllerList "github.com/khiki1995/todo/controllers/list"
)

func RouteList(route fiber.Router) {
	listRouter := route.Group("/list")

	listRouter.Get("/", controllerList.Get)
	listRouter.Get("/:id", controllerList.GetOne)
	listRouter.Post("/", controllerList.Create)
	listRouter.Delete("/:id", controllerList.Delete)
	listRouter.Put("/", controllerList.Update)
	listRouter.Get("/:id/items", controllerList.GetItems)
}
