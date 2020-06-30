package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostInput struct {
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
}

type Post struct {
	ID          string               `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string               `json:"title" bson:"title"`
	Description string               `json:"description" bson:"description"`
	CreatedAt   time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt" bson:"updatedAt"`
	Author      Author               `json:"author" bson:"author"`
	Comments    []primitive.ObjectID `json:"comments" bson:"comments"`
	Likes       []primitive.ObjectID `json:"likes" bson:"likes"`
}
