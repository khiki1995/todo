package routes

import (
	"github.com/gofiber/fiber/v2"
	controllerItem "github.com/khiki1995/todo/controllers/item"
)

func RouteListItem(route fiber.Router) {
	itemRouter := route.Group("/item")

	itemRouter.Get("/", controllerItem.Get)
	itemRouter.Get("/:id", controllerItem.GetOne)
	itemRouter.Post("/", controllerItem.Create)
	itemRouter.Delete("/:id", controllerItem.Delete)
	itemRouter.Put("/reorder", controllerItem.Reorder)
	itemRouter.Put("/", controllerItem.Update)
}
