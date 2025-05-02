package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo-contrib/session"
)

// checkLoginSession checks if a user session exists and renders the index page with their username if authenticated.
func checkLoginSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err == nil && sess != nil {
			username, ok := sess.Values["username"].(string)
			if ok && username != "" {
				return c.Render(http.StatusOK, "index", map[string]string{"Username": username})
			}
		}
		return next(c)
	}
}

// cacheControlMiddleware forces no caching for non-htmx requests.
func cacheControlMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("HX-Request") == "" {
			c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Response().Header().Set("Pragma", "no-cache")
			c.Response().Header().Set("Expires", "0")
		}
		return next(c)
	}
}
