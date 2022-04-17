package item

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/khiki1995/todo/config"
	"github.com/khiki1995/todo/helper"
	"github.com/khiki1995/todo/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func Create(ctx *fiber.Ctx) error {
	item := models.ToDoListItem{}

	if err := ctx.BodyParser(&item); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	errValid := helper.ValidateStruct(item)
	if errValid != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errValid)
	}

	todo := models.ToDoList{}
	item.CreatedOn = time.Now().UTC()
	err := config.Collections.ToDoList.FindOne(context.Background(), bson.M{"_id": item.ToDoListID}).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx.Status(fiber.StatusNotFound).JSON(fmt.Sprintf("list by (uuid = %s) does not exist", item.ToDoListID))
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	countItems, err := config.Collections.ToDoListItem.CountDocuments(context.Background(), bson.M{"todo_id": item.ToDoListID})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	// set to end of todolist
	item.Order = int(countItems)

	insertResult, err := config.Collections.ToDoListItem.InsertOne(context.Background(), item)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return helper.ResponseData(ctx, insertResult.InsertedID)
}

func GetOne(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	item := models.ToDoListItem{}

	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("Hex conversion: %s ", err))
	}

	err = config.Collections.ToDoListItem.FindOne(context.TODO(), bson.M{"_id": idHex}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx.Status(fiber.StatusNotFound).JSON(err.Error())
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return helper.ResponseData(ctx, item)
}

func Get(ctx *fiber.Ctx) error {
	toDoListItem := []models.ToDoListItem{}

	cursor, err := config.Collections.ToDoListItem.Find(context.TODO(), bson.M{})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if err := cursor.Err(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if err = cursor.All(context.TODO(), &toDoListItem); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	_ = cursor.Close(context.TODO())

	return helper.ResponseData(ctx, toDoListItem)
}

func Update(ctx *fiber.Ctx) error {
	item := models.ToDoListItem{}

	if err := ctx.BodyParser(&item); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	errValid := helper.ValidateStruct(item)
	if errValid != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errValid)
	}

	err := config.Collections.ToDoList.FindOne(context.TODO(), bson.M{"_id": item.ToDoListID}).Decode("")
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx.Status(fiber.StatusNotFound).JSON("not found by current todo_id")
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	update := bson.M{
		"$set": item,
	}

	result, err := config.Collections.ToDoListItem.UpdateOne(context.Background(), bson.M{"_id": item.ID}, update)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	if result.ModifiedCount == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON("not found")
	}

	return helper.ResponseData(ctx, item)
}

func Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	item := models.ToDoListItem{}
	err = config.Collections.ToDoListItem.FindOne(context.TODO(), bson.M{"_id": idHex}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ctx.Status(fiber.StatusNotFound).JSON(err.Error())
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	countItems, err := config.Collections.ToDoListItem.CountDocuments(context.Background(), bson.M{"todo_id": item.ToDoListID})
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return config.MI.Client.UseSession(context.Background(), func(sctx mongo.SessionContext) error {
		err := sctx.StartTransaction()
		if err != nil {
			return err
		}

		if countItems > 2 {
			err = reorderForItems(ctx, &models.ReorderRequest{
				ItemID:     item.ID,
				ToDoListID: item.ToDoListID,
				OrderFrom:  item.Order,
				OrderTo:    int(countItems) - 1,
			})
		}

		if err != nil {
			_ = sctx.AbortTransaction(sctx)
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		_, err = config.Collections.ToDoListItem.DeleteOne(context.TODO(), bson.M{"_id": idHex})
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		if err = sctx.CommitTransaction(sctx); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		return ctx.Status(fiber.StatusOK).JSON("")
	})
}

func Reorder(ctx *fiber.Ctx) error {
	req := models.ReorderRequest{}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	errValid := helper.ValidateStruct(req)
	if errValid != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(errValid)
	}

	item := models.ToDoListItem{}
	err := config.Collections.ToDoListItem.FindOne(context.TODO(), bson.M{"_id": req.ItemID}).Decode(&item)
	if err != nil {
		return err
	}

	countItems, err := config.Collections.ToDoListItem.CountDocuments(context.Background(), bson.M{"todo_id": item.ToDoListID})
	if err == mongo.ErrNoDocuments {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("list by (uuid = %s) does not exist", item.ToDoListID))
	} else if err != nil {
		log.Fatal(err)
	}
	if int64(req.OrderTo) >= countItems {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fmt.Sprintf("max order where can you move is  %d", countItems-1))
	}

	req.ToDoListID = item.ToDoListID
	req.OrderFrom = item.Order
	err = reorderForItems(ctx, &req)
	if err != nil {
		return err
	}

	return ctx.Status(fiber.StatusOK).JSON("successful")
}

func reorderForItems(ctx *fiber.Ctx, req *models.ReorderRequest) error {
	items := []models.ToDoListItem{}
	fmt.Println(req.ToDoListID.String())
	opts := options.Find().SetSort(bson.D{{"sort_num", 1}})
	cursor, err := config.Collections.ToDoListItem.Find(context.TODO(), bson.M{
		"todo_id": req.ToDoListID,
		"order": bson.M{
			"$gte": helper.GetMin(req.OrderFrom, req.OrderTo),
			"$lte": helper.GetMax(req.OrderFrom, req.OrderTo),
		},
	}, opts)

	if err != nil {
		return err
	}

	if err := cursor.Err(); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if err = cursor.All(context.TODO(), &items); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	if len(items) < 2 {
		return errors.New("two or more items required for swapping")
	}

	num := 1
	if req.OrderFrom > req.OrderTo {
		num = -1
	}

	return config.MI.Client.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		for _, item := range items {
			item.Order -= num
			if req.ItemID == item.ID {
				item.Order = req.OrderTo
			}

			_, err := config.Collections.ToDoListItem.UpdateOne(context.Background(), bson.M{"_id": item.ID}, bson.M{
				"$set": item,
			})
			if err != nil {
				_ = sessionContext.AbortTransaction(sessionContext)
				return ctx.Status(fiber.StatusInternalServerError).JSON(err.Error())
			}
		}

		if err = sessionContext.CommitTransaction(sessionContext); err != nil {
			return err
		}
		return nil
	})
}
