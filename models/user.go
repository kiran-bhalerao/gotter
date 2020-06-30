package models

import (
	"github.com/kiranbhalerao123/gotter/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        string               `json:"id,omitempty" bson:"_id,omitempty"`
	Email     string               `json:"email" bson:"email"`
	UserName  string               `json:"username" bson:"username"`
	Password  string               `json:"-" bson:"password,omitempty"`
	Posts     []primitive.ObjectID `json:"posts,omitempty" bson:"posts"`
	Following []primitive.ObjectID `json:"following,omitempty" bson:"following"`
	Followers []primitive.ObjectID `json:"followers,omitempty" bson:"followers"`
}

type Author struct {
	ID       string `json:"id,omitempty" bson:"_id,omitempty"`
	UserName string `json:"username" bson:"username"`
}

type SignupInputs struct {
	Email    string `json:"email" bson:"email" valid:"email"`
	UserName string `json:"username" bson:"username" valid:"length(3|30)"`
	Password string `json:"password" bson:"password,omitempty" valid:"length(6|30)"`
}

func (i SignupInputs) Validate() error {
	return utils.Validator(i)
}
