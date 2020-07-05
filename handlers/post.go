package handlers

import (
	"strconv"
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
		{"$match": bson.M{"author._id": userId}},
		{
			"$lookup": bson.M{
				"from": "comments",
				"let":  bson.M{"comments": "$comments"},
				"pipeline": bson.A{
					bson.M{"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$comments"}}}},
					bson.M{"$project": bson.M{
						"_id":       1,
						"message":   1,
						"post":      1,
						"user":      1,
						"createdAt": 1,
						"likes":     1,
						"count":     bson.M{"$size": "$likes"},
					}},
					bson.M{"$sort": bson.M{"count": -1}},
					bson.M{"$limit": limit},
					bson.M{"$project": bson.M{"count": 0}},
				},
				"as": "comments",
			},
		},
		{"$sort": bson.M{"createdAt": -1}},
		{"$skip": skip},
		{"$limit": limit},
	})

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	var posts []models.PostWithComment

	for cur.Next(c.Fasthttp) {
		// raw, err := cur.Current.Elements()
		var post models.PostWithComment
		err = cur.Decode(&post)

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

func (p PostHandler) HomeTimeline(c *fiber.Ctx) {
	limit := int64(10)
	page := int64(1)

	userId, err := primitive.ObjectIDFromHex(c.Params("userId"))
	if err != nil {
		userId = primitive.NilObjectID
	}

	// get limit from query
	lim, err := strconv.Atoi(c.Query("limit"))
	if err == nil {
		limit = int64(lim)
	}

	// get page from query
	pag, err := strconv.Atoi(c.Query("page"))
	if err == nil {
		page = int64(pag)
	}

	skip := (page - 1) * limit
	var query []primitive.M
	var cur *mongo.Cursor

	if userId != primitive.NilObjectID {
		// if userId is provided then get this user's followings
		// and with that followings get their posts

		query = []bson.M{
			{"$match": bson.M{"_id": userId}},
			{"$lookup": bson.M{
				"from": "users",
				"let":  bson.M{"following": "$following"},
				"pipeline": bson.A{
					bson.M{"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$following"}}}},
					bson.M{"$lookup": bson.M{
						"from": "posts",
						"let":  bson.M{"posts": "$posts"},
						"pipeline": bson.A{
							bson.M{"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$posts"}}}},
							bson.M{"$lookup": bson.M{
								"from": "comments",
								"let":  bson.M{"comments": "$comments"},
								"pipeline": bson.A{
									bson.M{"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$comments"}}}},
									bson.M{"$project": bson.M{
										"_id":       1,
										"message":   1,
										"post":      1,
										"user":      1,
										"createdAt": 1,
										"likes":     1,
										"count":     bson.M{"$size": "$likes"},
									}},
									bson.M{"$sort": bson.M{"count": -1}},
									bson.M{"$limit": limit},
									bson.M{"$project": bson.M{"count": 0}},
								},
								"as": "comments",
							}},
						},
						"as": "posts",
					}},
					bson.M{"$unwind": bson.M{
						"path":                       "$posts",
						"preserveNullAndEmptyArrays": true,
					}},
				},
				"as": "following",
			}},
			{"$unwind": bson.M{
				"path":                       "$following",
				"preserveNullAndEmptyArrays": true,
			}},
			{"$project": bson.M{"posts": "$following.posts"}},
			{"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": bson.A{"$posts", "$$ROOT"},
				},
			}},
			{"$project": bson.M{"posts": 0}},
			{"$sort": bson.M{"createdAt": -1}},
			{"$skip": skip},
			{"$facet": bson.M{
				"count": bson.A{bson.M{"$count": "count"}},
				"posts": bson.A{bson.M{"$limit": limit}},
			}},
			{"$project": bson.M{
				"count": bson.M{"$arrayElemAt": bson.A{"$count", 0}},
				"posts": 1,
			}},
			{"$project": bson.M{
				"count": "$count.count",
				"posts": 1,
			}},
		}

		cur, err = p.UserColl.Aggregate(c.Fasthttp, query)
	} else {
		// userId is not provided
		// get the latest posts from system

		query = []bson.M{
			{
				"$lookup": bson.M{
					"from": "comments",
					"let":  bson.M{"comments": "$comments"},
					"pipeline": bson.A{
						bson.M{"$match": bson.M{"$expr": bson.M{"$in": bson.A{"$_id", "$$comments"}}}},
						bson.M{"$project": bson.M{
							"_id":       1,
							"message":   1,
							"post":      1,
							"user":      1,
							"createdAt": 1,
							"likes":     1,
							"count":     bson.M{"$size": "$likes"},
						}},
						bson.M{"$sort": bson.M{"count": -1}},
						bson.M{"$limit": limit},
						bson.M{"$project": bson.M{"count": 0}},
					},
					"as": "comments",
				},
			},
			{"$sort": bson.M{"createdAt": -1}},
			{"$skip": skip},
			{"$facet": bson.M{
				"count": bson.A{bson.M{"$count": "count"}},
				"posts": bson.A{bson.M{"$limit": limit}},
			}},
			{"$project": bson.M{
				"count": bson.M{"$arrayElemAt": bson.A{"$count", 0}},
				"posts": 1,
			}},
			{"$project": bson.M{
				"count": "$count.count",
				"posts": 1,
			}},
		}

		cur, err = p.PostColl.Aggregate(c.Fasthttp, query)
	}

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}

	type Data struct {
		Count int32                    `json:"count"`
		Posts []models.PostWithComment `json:"posts"`
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
		Count: data[0].Count + int32(skip),
		Posts: data[0].Posts,
	})

	if err != nil {
		c.Status(fiber.StatusBadRequest).Send(err)
		return
	}
}
