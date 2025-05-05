package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo-contrib/session"
)

// requireLogout checks if a user session exists and renders the index page with their username if authenticated.
func requireLogout(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err == nil && sess != nil {
			username, ok := sess.Values["username"].(string)
			if ok && username != "" {
				return redirectIndex(c, map[string]string{"Username": username})
			}
		}
		return next(c)
	}
}
