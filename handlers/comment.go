package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber"
	conf "github.com/kiranbhalerao123/gotter/config"
	"github.com/kiranbhalerao123/gotter/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentHandler struct {
	CommentColl *mongo.Collection
	PostColl    *mongo.Collection
}

type CommentHandlerInterface interface {
	CommentPost(c *fiber.Ctx) interface{}
	UpdateComment(c *fiber.Ctx) interface{}
	DeleteComment(c *fiber.Ctx) interface{}
	LikeDislikeComment(c *fiber.Ctx) interface{}
	GetComment(c *fiber.Ctx) interface{}
}

/**
 * @Route /comment
 * @Body {postId: string, message: string}
 * @Mothod POST
 * @Protected ✔️
 */
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
		Likes:     []primitive.ObjectID{},
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

	if err := c.Status(fiber.StatusCreated).JSON(comment); err != nil {
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
	err = CH.CommentColl.FindOneAndUpdate(c.Fasthttp, filter, update, &conf.MongoOps.New).Decode(&comment)

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

func (CH CommentHandler) LikeDislikeComment(c *fiber.Ctx) {
	user := c.Locals("user").(models.User)

	userId, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	commentId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	var comment models.Comment
	// check whether the comment exists or not
	err = CH.CommentColl.FindOne(c.Fasthttp, bson.M{"_id": commentId}).Decode(&comment)
	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	// check whether the user already liked the comment
	err = CH.CommentColl.FindOne(c.Fasthttp, bson.M{"_id": commentId, "likes": bson.M{"$in": bson.A{userId}}}).Decode(&models.Comment{})

	var alreadyLiked bool

	if err == nil {
		alreadyLiked = true
	} else {
		if err.Error() != "mongo: no documents in result" {
			c.Status(fiber.StatusBadRequest).Send(err)
			return
		}
	}

	// update the comment doc
	filter := bson.M{"_id": commentId}
	update := bson.M{"$push": bson.M{"likes": userId}}

	if alreadyLiked {
		update = bson.M{"$pull": bson.M{"likes": userId}}
	}

	_, err = CH.CommentColl.UpdateOne(c.Fasthttp, filter, update)

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	message := "Comment Liked"
	if alreadyLiked {
		message = "Comment DisLiked"
	}

	c.Status(fiber.StatusOK).Send(message)
}

func (CH CommentHandler) GetComment(c *fiber.Ctx) {
	limit := int64(10)
	page := int64(1)

	// get limit from query
	lim, err := strconv.Atoi(c.Query("limit"))
	if err == nil {
		limit = int64(lim)
		if limit <= 0 {
			limit = 10 // set to default
		}
	}

	// get page from query
	pag, err := strconv.Atoi(c.Query("page"))
	if err == nil {
		page = int64(pag)
		if page <= 0 {
			page = 1 // set to default
		}
	}

	skip := (page - 1) * limit

	cur, err := CH.CommentColl.Aggregate(c.Fasthttp, []bson.M{
		{"$project": bson.M{
			"_id":       1,
			"message":   1,
			"post":      1,
			"user":      1,
			"createdAt": 1,
			"likes":     1,
			"count":     bson.M{"$size": "$likes"},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$skip": skip},
		{"$facet": bson.M{
			"count":    bson.A{bson.M{"$count": "count"}},
			"comments": bson.A{bson.M{"$limit": limit}},
		}},
		{"$project": bson.M{
			"count":    bson.M{"$arrayElemAt": bson.A{"$count", 0}},
			"comments": 1,
		}},
		{"$project": bson.M{
			"count":    "$count.count",
			"comments": 1,
		}},
	})

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	type Data struct {
		Count    int32            `json:"count"`
		Comments []models.Comment `json:"comments"`
	}

	var data []Data

	for cur.Next(c.Fasthttp) {
		var d Data
		err = cur.Decode(&d)

		if err != nil {
			c.Status(fiber.StatusBadRequest).Send(err)
			return
		}
		data = append(data, d)
	}

	if err := cur.Err(); err != nil {
		if err != nil {
			c.Status(fiber.StatusBadRequest).Send(err)
			return
		}
	}

	// Close the cursor once finished
	cur.Close(c.Fasthttp)
	err = c.Status(fiber.StatusOK).JSON(Data{
		Count:    data[0].Count + int32(skip),
		Comments: data[0].Comments,
	})

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}
}
