package handlers

import (
	"github.com/gofiber/fiber"
	. "github.com/kiranbhalerao123/gotter/config"
	"github.com/kiranbhalerao123/gotter/models"
	"github.com/kiranbhalerao123/gotter/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserHandlersInterface interface {
	GetUser(ctx *fiber.Ctx) interface{}
	UpdateUser(ctx *fiber.Ctx) interface{}
	FollowUnFollowUser(c *fiber.Ctx) interface{}
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

	if err := c.Status(200).JSON(usr); err != nil {
		c.Status(500).Send(err)
		return
	}
}

func (u UserHandler) UpdateUser(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	var inputs models.UpdateInputs
	var username = user.UserName
	var password = user.Password

	userId, err := primitive.ObjectIDFromHex(user.ID)

	// the provided ID might be invalid ObjectID
	if err != nil {
		c.Status(400).Send(err)
		return
	}

	if err := c.BodyParser(&inputs); err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
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

	var updatedUser models.User

	err = u.UserColl.FindOneAndUpdate(c.Fasthttp, filter, update, &MongoOps.New).Decode(&updatedUser)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	if err := c.Status(200).JSON(updatedUser); err != nil {
		c.Status(500).Send(err)
		return
	}
}

/**
 * @Params /:id
 *  - another users id
 * @Mothod POST
 * @Protected ✔️
 */
func (u UserHandler) FollowUnFollowUser(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	currentUserId, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	// I want to follow the another user, I'm following to anotherUserId
	anotherUserId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	// check the user exists or not
	err = u.UserColl.FindOne(c.Fasthttp, bson.M{"_id": anotherUserId}).Decode(&models.User{})
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	// check for already following
	err = u.UserColl.FindOne(c.Fasthttp, bson.M{"_id": anotherUserId, "followers": bson.M{"$in": bson.A{currentUserId}}}).Decode(&models.User{})

	if err != nil && err.Error() != "mongo: no documents in result" {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	alreadyFollowing := err == nil || err.Error() != "mongo: no documents in result"

	// follow the user
	currentUserUpdate := bson.M{"$push": bson.M{"following": anotherUserId}}
	anotherUserUpdate := bson.M{"$push": bson.M{"followers": currentUserId}}

	if alreadyFollowing {
		// unfollow the user
		currentUserUpdate = bson.M{"$pull": bson.M{"following": anotherUserId}}
		anotherUserUpdate = bson.M{"$pull": bson.M{"followers": currentUserId}}
	}

	// follow/unfollow the user
	_, err = u.UserColl.UpdateOne(c.Fasthttp, bson.M{"_id": currentUserId}, currentUserUpdate)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	// add/remove from another users followers[]
	_, err = u.UserColl.UpdateOne(c.Fasthttp, bson.M{"_id": anotherUserId}, anotherUserUpdate)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	message := "Followed the user"
	if alreadyFollowing {
		message = "UnFollowed the user"
	}

	if err := c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     message,
		"isFollowing": !alreadyFollowing,
	}); err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}
}
