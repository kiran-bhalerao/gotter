package models

import (
	"time"

	"github.com/kiranbhalerao123/gotter/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostInput struct {
	Title       string `json:"title" bson:"title" valid:"length(3|30)"`
	Description string `json:"description" bson:"description" valid:"length(3|300)"`
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

type PostWithComment struct {
	ID          string               `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string               `json:"title" bson:"title"`
	Description string               `json:"description" bson:"description"`
	CreatedAt   time.Time            `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt" bson:"updatedAt"`
	Author      Author               `json:"author" bson:"author"`
	Comments    []Comment            `json:"comments" bson:"comments"`
	Likes       []primitive.ObjectID `json:"likes" bson:"likes"`
}

func (i PostInput) Validate() error {
	return utils.Validator(i)
}
