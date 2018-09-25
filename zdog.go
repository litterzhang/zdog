package main

import (
	"zdog/render"
	"zdog/handler"
	"zdog/model"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func route(e *echo.Echo) *echo.Echo {
	// setup handler
	h := handler.New()

	// set static resource
	e.Static("/favicon.ico", "static/favicon.ico")
	e.Static("/static", "static")

	// set route
	e.GET("/", h.Index)
	e.GET("/index", h.Index)

	// set post route
	e.POST("/login", h.Login)
	e.GET("/login", h.GetLogin)
	e.GET("/logout", h.Logout)

	e.GET("/:route", h.RouteRender)

	e.HTTPErrorHandler = handler.ErrorHandler
	return e
}

func setup() *echo.Echo {
	e := echo.New()
	// middleware
	e.Use(middleware.Logger())

	// send render
	t := render.New()
	e.Renderer = t
	return e
}

func close() {
	model.CloseDb()
}

func main() {
	model.OpenDb()
	defer close()

	e := setup()
	e = route(e)
	e.Start("127.0.0.1:8888")
}
