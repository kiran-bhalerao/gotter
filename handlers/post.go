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

type PostHandlerIntreface interface {
	CreatePost(c *fiber.Ctx) interface{}
	UpdatePost(c *fiber.Ctx) interface{}
	DeletePost(c *fiber.Ctx) interface{}
	LikeDislikePost(c *fiber.Ctx) interface{}
	HomeTimeline(c *fiber.Ctx) interface{}
	UserTimeline(c *fiber.Ctx) interface{}
}

type PostHandler struct {
	PostColl    *mongo.Collection
	UserColl    *mongo.Collection
	CommentColl *mongo.Collection
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
		Likes:       []primitive.ObjectID{},
		Author: models.Author{
			ID:       user.ID,
			UserName: user.UserName,
		},
	}

	insertionResult, err := p.PostColl.InsertOne(c.Fasthttp, post)

	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
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
		return
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
		return
	}

	filter := bson.M{"_id": postId, "author._id": user.ID}
	deleteResult, e := p.PostColl.DeleteOne(c.Fasthttp, filter)

	if e != nil || deleteResult.DeletedCount < 1 {
		c.Status(fiber.StatusInternalServerError).Send("Unable to delete post")
		return
	}

	// pull out postId from users collection
	filter = bson.M{"email": user.Email}
	update := bson.M{"$pull": bson.M{"posts": postId}}

	_, err = p.UserColl.UpdateOne(c.Fasthttp, filter, update)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	// delete all comments associated with this post
	_, err = p.CommentColl.DeleteMany(c.Fasthttp, bson.M{"post": postId})
	if err != nil {
		c.Status(fiber.StatusInternalServerError).Send(err)
		return
	}

	c.Status(fiber.StatusOK).Send("Post deleted successfully")
}

func (P PostHandler) LikeDislikePost(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	userId, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	postId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	var post models.Post
	// check whether the post exist or not
	err = P.PostColl.FindOne(c.Fasthttp, bson.M{"_id": postId}).Decode(&post)

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	err = P.PostColl.FindOne(c.Fasthttp, bson.M{"_id": postId, "likes": bson.M{"$in": bson.A{userId}}}).Decode(&models.User{})

	if err != nil && err.Error() != "mongo: no documents in result" {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	// inline
	// notLikedYet := err != nil && err.Error() == "mongo: no documents in result"
	// but i prefer more explicit way

	var notLikedYet bool
	if err != nil && err.Error() == "mongo: no documents in result" {
		notLikedYet = true
	} else {
		notLikedYet = false
	}

	filter := bson.M{"_id": postId}
	update := bson.M{"$pull": bson.M{"likes": userId}}

	if notLikedYet {
		update = bson.M{"$push": bson.M{"likes": userId}}
	}

	_, err = P.PostColl.UpdateOne(c.Fasthttp, filter, update)

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	message := "Post DisLiked"
	if notLikedYet {
		message = "Post Liked"
	}

	c.Status(fiber.StatusOK).Send(message)
}

func (p PostHandler) UserTimeline(c *fiber.Ctx) {
	userId := c.Params("userId")
	limit := int64(10)
	page := int64(1)
	skip := (page - 1) * limit

	// get posts of the userId
	cur, err := p.PostColl.Aggregate(c.Fasthttp, []bson.M{
		{
			"$match": bson.M{
				"author._id": userId,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "comments",
				"localField":   "comments",
				"foreignField": "_id",
				"as":           "comments",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$comments",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$group": bson.M{
				"_id":         "$_id",
				"title":       bson.M{"$first": "$title"},
				"description": bson.M{"$first": "$description"},
				"createdAt":   bson.M{"$first": "$createdAt"},
				"author":      bson.M{"$first": "$author"},
				"likes":       bson.M{"$first": "$likes"},
				"comments":    bson.M{"$addToSet": "$comments"},
			},
		},
		{"$sort": bson.M{"createdAt": -1}},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	})

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	type Post struct {
		ID          string               `json:"id,omitempty" bson:"_id,omitempty"`
		Title       string               `json:"title" bson:"title"`
		Description string               `json:"description" bson:"description"`
		CreatedAt   time.Time            `json:"createdAt" bson:"createdAt"`
		UpdatedAt   time.Time            `json:"updatedAt" bson:"updatedAt"`
		Author      models.Author        `json:"author" bson:"author"`
		Comments    []models.Comment     `json:"comments" bson:"comments"`
		Likes       []primitive.ObjectID `json:"likes" bson:"likes"`
	}

	var posts []Post

	for cur.Next(c.Fasthttp) {
		var post Post
		err := cur.Decode(&post)

		if err != nil {
			c.Status(fiber.StatusBadRequest).Send(err)
			return
		}
		posts = append(posts, post)
	}

	if err := cur.Err(); err != nil {
		if err != nil {
			c.Status(fiber.StatusBadRequest).Send(err)
			return
		}
	}

	// Close the cursor once finished
	cur.Close(c.Fasthttp)
	err = c.Status(fiber.StatusOK).JSON(fiber.Map{
		"posts": posts,
		"count": len(posts),
	})

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}
}
