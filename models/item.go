package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Priority string

const (
	PriorityLowest  Priority = "lowest"
	PriorityLow     Priority = "low"
	PriorityMedium  Priority = "medium"
	PriorityHigh    Priority = "high"
	PriorityHighest Priority = "highest"
)

type ToDoListItem struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ToDoListID  primitive.ObjectID `json:"todo_id" bson:"todo_id" validate:"required"`
	Description string             `json:"description,omitempty" bson:"description,omitempty" validate:"required"`
	Priority    Priority           `json:"priority,omitempty" bson:"priority,omitempty" validate:"required,oneof='lowest' 'low' 'medium' 'high' 'highest'"`
	Order       int                `json:"order" bson:"order"`
	CreatedOn   time.Time          `json:"created_on,omitempty" bson:"created_on,omitempty"`
}

type ReorderRequest struct {
	ItemID     primitive.ObjectID `json:"id" validate:"required"`
	ToDoListID primitive.ObjectID `json:"todolist_id,omitempty"`
	OrderFrom  int                `json:"order_from,omitempty" validate:"min=0"`
	OrderTo    int                `json:"order_to,omitempty" validate:"min=0"`
}
