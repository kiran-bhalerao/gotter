package router

import (
	"github.com/gofiber/fiber"
	. "github.com/kiranbhalerao123/gotter/config"
	. "github.com/kiranbhalerao123/gotter/handlers"
	. "github.com/kiranbhalerao123/gotter/middlewares"
)

func SetupRouter(app *fiber.App) {
	// Router Setup
	router := app.Group("/api/v1")

	// Auth Routes
	_authHandler := AuthHandler{UsersColl: Mongo.DB.Collection("users")}
	router.Post("/signup", _authHandler.Signup)
	router.Post("/login", _authHandler.Login)

	// User Routes
	_userHandler := UserHandler{UserColl: Mongo.DB.Collection("users")}
	router.Get("/user", WithGuard, WithUser, _userHandler.GetUser)
	router.Put("/user", WithGuard, WithUser, _userHandler.UpdateUser)
	router.Post("/user/:id", WithGuard, WithUser, _userHandler.FollowUnFollowUser)

	// Post Routes
	_postHandler := PostHandler{
		UserColl:    Mongo.DB.Collection("users"),
		PostColl:    Mongo.DB.Collection("posts"),
		CommentColl: Mongo.DB.Collection("comments"),
	}
	router.Post("/post", WithGuard, WithUser, _postHandler.CreatePost)
	router.Put("/post/:id", WithGuard, WithUser, _postHandler.UpdatePost)
	router.Delete("/post/:id", WithGuard, WithUser, _postHandler.DeletePost)
	router.Post("/post/:id", WithGuard, WithUser, _postHandler.LikeDislikePost)
	router.Get("/post/timeline/user/:userId", _postHandler.UserTimeline)  // another users userId
	router.Get("/post/timeline/home/:userId?", _postHandler.HomeTimeline) // current users userId (optional)

	// Comment Routes
	_commentHandler := CommentHandler{
		CommentColl: Mongo.DB.Collection("comments"),
		PostColl:    Mongo.DB.Collection("posts"),
	}
	router.Get("/comment", _commentHandler.GetComment)
	router.Post("/comment", WithGuard, WithUser, _commentHandler.CommentPost)
	router.Put("/comment/:id", WithGuard, WithUser, _commentHandler.UpdateComment)
	router.Delete("/comment/:id", WithGuard, WithUser, _commentHandler.DeleteComment)
	router.Post("/comment/:id", WithGuard, WithUser, _commentHandler.LikeDislikeComment)
}
