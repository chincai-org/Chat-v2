package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo-contrib/session"
)

const StatusUnprocessableEntity = 422

// renderIndex renders the index page.
func renderIndex(c echo.Context) error {
	var username string

	sess, err := session.Get("session", c)
	if err == nil && sess != nil {
		name, ok := sess.Values["username"].(string)
		if ok && name != "" {
			username = name
		}
	}

	// Initial request
	if c.Request().Header.Get("HX-Request") == "" {
		return c.Render(http.StatusOK, "index", map[string]string{"Username": username})
	}

	return c.Render(http.StatusOK, "main", map[string]string{"Username": username})
}

// renderSignup renders the signup page.
func renderSignup(c echo.Context) error {
	return c.Render(http.StatusOK, "signup", nil)
}

// renderSignin renders the signin page.
func renderSignin(c echo.Context) error {
	return c.Render(http.StatusOK, "signin", nil)
}

// signupValidator handles the signup form validation.
func signupValidator(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm-password")

	formData := newFormData()
	formData.Values["username"] = username
	formData.Values["password"] = password
	formData.Values["confirmPassword"] = confirmPassword

	existUsername := existUsername(username)
	if len(username) < 3 {
		formData.Errors["username"] = "Username must be at least 3 characters long"
	} else if len(username) > 15 {
		formData.Errors["username"] = "Username must be at most 15 characters long"
	} else if existUsername {
		formData.Errors["username"] = "Username already exists"
	} else if invalidUsername(username) {
		formData.Errors["username"] = "Username can only contain alphabets, numbers, underscore (_) and dash (-)"
	} else if len(password) < 6 {
		formData.Errors["password"] = "Password must be at least 6 characters long"
	} else if len(password) > 30 {
		formData.Errors["password"] = "Password must be at most 30 characters long"
	} else if confirmPassword != password {
		formData.Errors["confirmPassword"] = "Password does not match"
	}

	if len(formData.Errors) > 0 {
		return c.Render(StatusUnprocessableEntity, "signup", formData)
	}

	user := newUser(username, password)
	if err := user.save(); err != nil {
        formData.Errors["username"] = "Username taken while processing"
        return c.Render(StatusUnprocessableEntity, "signup", formData)
    }
	return user.login(c)
}

// signinValidator handles the signin form validation.
func signinValidator(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	formData := newFormData()
	formData.Values["username"] = username
	formData.Values["password"] = password

	existUsername := existUsername(username)
	if !existUsername {
		formData.Errors["username"] = "Username does not exist"
		return c.Render(StatusUnprocessableEntity, "signin", formData)
	}

	readPassword, err := getPassword(username)
	if err != nil {
		formData.Errors[""] = "An unexpected error occurred. Please try again."
		return c.Render(http.StatusInternalServerError, "signin", formData)
	}

	if readPassword != password {
		formData.Errors["password"] = "Password does not match"
		return c.Render(StatusUnprocessableEntity, "signin", formData)
	}

	user := newUser(username, password)
	return user.login(c)
}

// logoutHandler handles logging out by clearing the session.
func logoutHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err == nil {
		sess.Options.MaxAge = -1 // Delete the session
		sess.Save(c.Request(), c.Response())
	}

	return redirectIndex(c, nil)
}

func redirectIndex(c echo.Context, data any) error{
	if c.Request().Header.Get("HX-Request") != "" {
		c.Response().Header().Set("HX-Push", "/")
		return c.Render(http.StatusOK, "main", data)
	}
	return c.Redirect(http.StatusSeeOther, "/")
}
