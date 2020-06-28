package main

import (
	c "github.com/kiranbhalerao123/gotter/config"
	h "github.com/kiranbhalerao123/gotter/handlers"
	m "github.com/kiranbhalerao123/gotter/middlewares"
)

func init() {
	router := App.Group("/api/v1")

	a := h.AuthHandler{UsersColl: c.Mongo.DB.Collection("users")}
	router.Post("/signup", a.Signup)
	router.Post("/login", a.Login)

	u := h.UserHandler{UserColl: c.Mongo.DB.Collection("users")}
	router.Get("/user", m.WithGuard, m.WithUser, u.GetUser)
	router.Put("/user", m.WithGuard, m.WithUser, u.UpdateUser)
	router.Delete("/user", m.WithGuard, m.WithUser, u.DeleteUser)

	p := h.PostHandler{UserColl: c.Mongo.DB.Collection("users"), PostColl: c.Mongo.DB.Collection("posts")}
	router.Post("/post", m.WithGuard, m.WithUser, p.CreatePost)
	router.Put("/post/:id", m.WithGuard, m.WithUser, p.UpdatePost)
	router.Delete("/post/:id", m.WithGuard, m.WithUser, p.DeletePost)
	router.Delete("/post", m.WithGuard, m.WithUser, p.DeleteAllPost)
}
