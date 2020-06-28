package handlers

import (
	"github.com/gofiber/fiber"
	"github.com/kiranbhalerao123/gotter/models"
	"github.com/kiranbhalerao123/gotter/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserHandlersInterface interface {
	GetUser(ctx *fiber.Ctx) interface{}
	UpdateUser(ctx *fiber.Ctx) interface{}
	DeleteUser(ctx *fiber.Ctx) interface{}
}

type UserHandler struct {
	UserColl *mongo.Collection
}

func (u UserHandler) GetUser(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	userId, err := primitive.ObjectIDFromHex(user.ID)

	// the provided ID might be invalid ObjectID
	if err != nil {
		c.Status(400).Send(err)
		return
	}

	filter := bson.D{{Key: "_id", Value: userId}}

	var usr models.User
	err = u.UserColl.FindOne(c.Fasthttp, filter).Decode(&usr)

	if err != nil {
		c.Status(400).Send(err)
		return
	}

	if err := c.Status(201).JSON(usr); err != nil {
		c.Status(500).Send(err)
		return
	}
}

func (u UserHandler) UpdateUser(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	var inputs models.User
	var username = user.UserName
	var password = user.Password

	userId, err := primitive.ObjectIDFromHex(user.ID)

	// the provided ID might be invalid ObjectID
	if err != nil {
		c.Status(400).Send(err)
		return
	}

	if err := c.BodyParser(&inputs); err != nil {
		if err := c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid data"}); err != nil {
			c.Status(fiber.StatusInternalServerError).Send(err)
		}
		return
	}

	if inputs.UserName != "" {
		username = inputs.UserName
	}

	if inputs.Password != "" {
		hashPassword := utils.Password{Password: inputs.Password}.Hash()

		password = hashPassword
	}

	filter := bson.D{{Key: "_id", Value: userId}}
	update := bson.M{"$set": bson.M{"username": username, "password": password}}
	// update := bson.D{{Key: "$set", Value: bson.M{"username": username, "password": password}}} ✔️

	// Create an instance of an options and set the desired options
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	var updatedUser models.User

	err = u.UserColl.FindOneAndUpdate(c.Fasthttp, filter, update, &opt).Decode(&updatedUser)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	if err := c.Status(201).JSON(updatedUser); err != nil {
		c.Status(500).Send(err)
		return
	}
}

func (u UserHandler) DeleteUser(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	userId, err := primitive.ObjectIDFromHex(user.ID)

	// the provided ID might be invalid ObjectID
	if err != nil {
		c.Status(400).Send(err)
		return
	}

	filter := bson.M{"_id": userId}

	result, errors := u.UserColl.DeleteOne(c.Fasthttp, filter)

	if result.DeletedCount < 1 || errors != nil {
		c.Status(400).Send("Unable to delete user")
		return
	}

	c.Send("User deleted successfully")
}
