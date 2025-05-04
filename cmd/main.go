package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/echo-contrib/session"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(session.Middleware(sessionStore))
	e.Use(cacheControlMiddleware)

	e.Renderer = newTemplate()
	e.Static("/static/images", "images")
	e.Static("/static/css", "css")

	e.GET("/", renderIndex)
	e.GET("/signup", renderSignup, requireLogout)
	e.GET("/signin", renderSignin, requireLogout)

	e.POST("/signup-validator", signupValidator, requireLogout)
	e.POST("/signin-validator", signinValidator, requireLogout)
	e.POST("/logout", logoutHandler)

	e.Logger.Fatal(e.Start(":8000"))
}

