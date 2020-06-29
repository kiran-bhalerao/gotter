package handlers

import (
	"time"

	"github.com/gofiber/fiber"
	"github.com/kiranbhalerao123/gotter/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommentHandler struct {
	CommentColl *mongo.Collection
	PostColl    *mongo.Collection
}

type CommentHandlerInterface interface {
	CommentPost(c *fiber.Ctx) interface{}
	UpdateComment(c *fiber.Ctx) interface{}
	DeleteComment(c *fiber.Ctx) interface{}
}

func (CH CommentHandler) CommentPost(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	body := new(struct {
		PostId  string
		Message string
	})

	if e := c.BodyParser(body); e != nil {
		c.Status(fiber.StatusBadRequest).Send(e)
		return
	}

	postId, err := primitive.ObjectIDFromHex(body.PostId)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	// check if post is available
	var post models.Post
	e := CH.PostColl.FindOne(c.Fasthttp, bson.M{"_id": postId}).Decode(&post)
	if e != nil {
		c.Status(fiber.StatusBadRequest).Send(e)
		return
	}

	comment := models.Comment{
		Message:   body.Message,
		CreatedAt: time.Now(),
		Post:      postId,
		User:      models.Author{ID: user.ID, UserName: user.UserName},
	}

	// create comment
	insertedResult, err := CH.CommentColl.InsertOne(c.Fasthttp, comment)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	comment.ID = insertedResult.InsertedID.(primitive.ObjectID).Hex()

	// add comment to post
	filter := bson.M{"_id": postId}
	update := bson.M{"$push": bson.M{"comments": insertedResult.InsertedID}}
	_, err = CH.PostColl.UpdateOne(c.Fasthttp, filter, update)

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	if err := c.Status(fiber.StatusOK).JSON(comment); err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}
}

func (CH CommentHandler) UpdateComment(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)
	body := new(struct{ Comment string })

	if err := c.BodyParser(body); err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	commentId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	var comment models.Comment

	filter := bson.M{"_id": commentId, "user._id": user.ID}
	update := bson.M{"$set": bson.M{"message": body.Comment}}
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	err = CH.CommentColl.FindOneAndUpdate(c.Fasthttp, filter, update, &opt).Decode(&comment)

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	if err := c.Status(fiber.StatusOK).JSON(comment); err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}
}

func (CH CommentHandler) DeleteComment(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	commentId, err := primitive.ObjectIDFromHex(c.Params("id"))

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	var comment models.Comment

	e := CH.CommentColl.FindOneAndDelete(c.Fasthttp, bson.M{"_id": commentId, "user._id": user.ID}).Decode(&comment)

	if e != nil {
		c.Status(fiber.StatusInternalServerError).Send(e)
		return
	}

	// pull out commentId from post collection's Comment[]
	filter := bson.M{"_id": comment.Post}
	update := bson.M{"$pull": bson.M{"comments": commentId}}

	_, err = CH.PostColl.UpdateOne(c.Fasthttp, filter, update)

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}
	c.Status(fiber.StatusOK).Send("Comment deleted successfully")
}
