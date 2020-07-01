package router

import (
	"github.com/gofiber/fiber"
	conf "github.com/kiranbhalerao123/gotter/config"
	handle "github.com/kiranbhalerao123/gotter/handlers"
	middle "github.com/kiranbhalerao123/gotter/middlewares"
)

func SetupRouter(app *fiber.App) {
	// Router Setup
	router := app.Group("/api/v1")

	// Auth Routes
	a := handle.AuthHandler{UsersColl: conf.Mongo.DB.Collection("users")}
	router.Post("/signup", a.Signup)
	router.Post("/login", a.Login)

	// User Routes
	u := handle.UserHandler{UserColl: conf.Mongo.DB.Collection("users")}
	router.Get("/user", middle.WithGuard, middle.WithUser, u.GetUser)
	router.Put("/user", middle.WithGuard, middle.WithUser, u.UpdateUser)
	router.Post("/user/:id", middle.WithGuard, middle.WithUser, u.FollowUnFollowUser)

	// Post Routes
	p := handle.PostHandler{
		UserColl:    conf.Mongo.DB.Collection("users"),
		PostColl:    conf.Mongo.DB.Collection("posts"),
		CommentColl: conf.Mongo.DB.Collection("comments"),
	}
	router.Post("/post", middle.WithGuard, middle.WithUser, p.CreatePost)
	router.Put("/post/:id", middle.WithGuard, middle.WithUser, p.UpdatePost)
	router.Delete("/post/:id", middle.WithGuard, middle.WithUser, p.DeletePost)
	router.Post("/post/:id", middle.WithGuard, middle.WithUser, p.LikeDislikePost)

	// Comment Routes
	c := handle.CommentHandler{
		CommentColl: conf.Mongo.DB.Collection("comments"),
		PostColl:    conf.Mongo.DB.Collection("posts"),
	}
	router.Post("/comment", middle.WithGuard, middle.WithUser, c.CommentPost)
	router.Put("/comment/:id", middle.WithGuard, middle.WithUser, c.UpdateComment)
	router.Delete("/comment/:id", middle.WithGuard, middle.WithUser, c.DeleteComment)
}
