package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID        string               `json:"id,omitempty" bson:"_id,omitempty"`
	Message   string               `json:"message" bson:"message"`
	Post      primitive.ObjectID   `json:"post" bson:"post"`
	User      Author               `json:"user" bson:"user"`
	CreatedAt time.Time            `json:"createdAt"`
	Likes     []primitive.ObjectID `json:"likes" bson:"likes"`
}
