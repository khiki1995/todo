package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ToDoList struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title,omitempty" bson:"title,omitempty" validate:"required"`
	Description string             `json:"description,omitempty" bson:"description,omitempty" validate:"required"`
	Color       string             `json:"color,omitempty" bson:"color,omitempty" validate:"required,hexcolor"`
}
