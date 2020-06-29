package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber"
	"github.com/kiranbhalerao123/gotter/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostHandlerIntreface interface {
	CreatePost(c *fiber.Ctx) interface{}
	UpdatePost(c *fiber.Ctx) interface{}
	DeletePost(c *fiber.Ctx) interface{}
	DeleteAllPost(c *fiber.Ctx) interface{}
}

type PostHandler struct {
	PostColl *mongo.Collection
	UserColl *mongo.Collection
}

func (p PostHandler) CreatePost(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	userId, e := primitive.ObjectIDFromHex(user.ID)

	if e != nil {
		c.Status(fiber.StatusBadRequest).Send(e)
		return
	}

	var inputs models.PostInput

	if err := c.BodyParser(&inputs); err != nil {
		if err := c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid Inputs"}); err != nil {
			c.Status(fiber.StatusInternalServerError).Send(err)
			return
		}
	}

	post := models.Post{
		Title:       inputs.Title,
		Description: inputs.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Comments:    []primitive.ObjectID{},
		Author: models.Author{
			ID:       user.ID,
			UserName: user.UserName,
		},
	}

	insertionResult, err := p.PostColl.InsertOne(c.Fasthttp, post)

	if err != nil {
		log.Fatal(err)
	}

	// update the users collection, put post id inside posts[]
	filter := bson.M{"_id": userId}
	update := bson.M{"$push": bson.M{"posts": insertionResult.InsertedID}}

	_, err = p.UserColl.UpdateOne(c.Fasthttp, filter, update)

	if err != nil {
		// rollback the post insertion
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	err = p.PostColl.FindOne(c.Fasthttp, bson.M{"_id": insertionResult.InsertedID}).Decode(&post)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	if err := c.Status(fiber.StatusCreated).JSON(post); err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
	}
}

func (p PostHandler) UpdatePost(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	postId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
	}

	var inputs models.PostInput

	if err := c.BodyParser(&inputs); err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	filter := bson.M{"_id": postId, "author._id": user.ID}
	update := bson.M{"$set": bson.M{"title": inputs.Title, "description": inputs.Description}}

	// update check for empty title
	if inputs.Title == "" {
		update = bson.M{"$set": bson.M{"description": inputs.Description}}
	}

	// update check for empty description
	if inputs.Description == "" {
		update = bson.M{"$set": bson.M{"title": inputs.Title}}
	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	var post models.Post

	err = p.PostColl.FindOneAndUpdate(c.Fasthttp, filter, update, &opt).Decode(&post)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	if err := c.Status(201).JSON(post); err != nil {
		c.Status(500).Send(err)
		return
	}
}

func (p PostHandler) DeletePost(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	postId, err := primitive.ObjectIDFromHex(c.Params("id"))

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
	}

	filter := bson.M{"_id": postId, "author._id": user.ID}
	deleteResult, e := p.PostColl.DeleteOne(c.Fasthttp, filter)

	if e != nil || deleteResult.DeletedCount < 1 {
		c.Status(fiber.StatusInternalServerError).Send("Unable to delete post")
	}

	// pull out postId from users collection
	filter = bson.M{"email": user.Email}
	update := bson.M{"$pull": bson.M{"posts": postId}}

	_, err = p.UserColl.UpdateOne(c.Fasthttp, filter, update)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	c.Status(fiber.StatusOK).Send("Post deleted successfully")
}

func (p PostHandler) DeleteAllPost(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	filter := bson.M{"author._id": user.ID}

	deleteResult, err := p.PostColl.DeleteMany(c.Fasthttp, filter)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	if deleteResult.DeletedCount < 1 {
		c.Status(fiber.StatusNotModified).Send("Unable to delete any posts")
		return
	}

	// empty posts[] from users collection
	filter = bson.M{"email": user.Email}
	update := bson.M{"$set": bson.M{"posts": []models.Post{}}}

	_, err = p.UserColl.UpdateOne(c.Fasthttp, filter, update)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	c.Status(fiber.StatusOK).Send("Posts deleted successfully")
}
