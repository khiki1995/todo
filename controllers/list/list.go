package list

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/khiki1995/todo/config"
	"github.com/khiki1995/todo/helper"
	"github.com/khiki1995/todo/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Create(ctx *fiber.Ctx) error {
	list := models.ToDoList{}

	if err := ctx.BodyParser(&list); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	errValid := helper.ValidateStruct(list)
	if errValid != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errValid)
	}

	insertResult, err := config.Collections.ToDoList.InsertOne(context.Background(), list)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return helper.ResponseData(ctx, insertResult.InsertedID)
}

func GetOne(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	item := models.ToDoList{}

	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	err = config.Collections.ToDoList.FindOne(context.TODO(), bson.M{"_id": idHex}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx.Status(fiber.StatusNotFound).JSON(err.Error())
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return helper.ResponseData(ctx, item)
}

func Get(ctx *fiber.Ctx) error {
	toDoList := []models.ToDoList{}

	cursor, err := config.Collections.ToDoList.Find(context.TODO(), bson.M{})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if err := cursor.Err(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if err = cursor.All(context.TODO(), &toDoList); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	_ = cursor.Close(context.TODO())

	return helper.ResponseData(ctx, toDoList)
}

func GetItems(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	items := []models.ToDoListItem{}
	cursor, err := config.Collections.ToDoListItem.Find(context.TODO(), bson.M{"todo_id": idHex})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if err := cursor.Err(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if err = cursor.All(context.TODO(), &items); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	_ = cursor.Close(context.TODO())

	return helper.ResponseData(ctx, items)
}

func Update(ctx *fiber.Ctx) error {
	list := models.ToDoList{}

	if err := ctx.BodyParser(&list); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	errValid := helper.ValidateStruct(list)
	if errValid != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errValid)
	}

	update := bson.M{
		"$set": list,
	}
	result, err := config.Collections.ToDoList.UpdateOne(context.Background(), bson.M{"_id": list.ID}, update)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	if result.ModifiedCount == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON("not found")
	}

	return helper.ResponseData(ctx, list)
}

func Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	return config.MI.Client.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		result, err := config.Collections.ToDoList.DeleteOne(context.TODO(), bson.M{"_id": idHex})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}
		if result.DeletedCount == 0 {
			return ctx.Status(fiber.StatusNotFound).JSON("")
		}

		items := []models.ToDoListItem{}

		cursor, err := config.Collections.ToDoListItem.Find(context.TODO(), bson.M{"todo_id": idHex})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		if err := cursor.Err(); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		if err = cursor.All(context.TODO(), &items); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		for _, item := range items {
			_, err = config.Collections.ToDoListItem.DeleteOne(context.TODO(), bson.M{"_id": item.ID})
			if err != nil {
				_ = sessionContext.AbortTransaction(sessionContext)
				return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
			}
		}

		if err = sessionContext.CommitTransaction(sessionContext); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		return ctx.Status(fiber.StatusOK).JSON("")
	})
}
