package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
}

type User struct {
	Username string
	Password string
}

func newUser(name, password string) User {
	return User{
		Username: name,
		Password: password,
	}
}

type Users = []User

type Data struct {
	Users Users
}

func (d *Data) existUsername(username string) (int, bool) {
	for i, user := range d.Users {
		if user.Username == username {
			return i, true
		}
	}
	return 0, false
}
func (d *Data) invalidUsername(username string) bool {
	for _, char := range username {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && char != '_' && char != '-' {
			return true
		}
	}
	return false
}

func newData() Data {
	return Data{
		Users: []User{},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	page := newPage()
	e.Renderer = newTemplate()
	e.Static("/static/images", "images")
	e.Static("/static/css", "css")

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.GET("/signup", func(c echo.Context) error {
		return c.Render(200, "signup", page)
	})

	e.GET("/signin", func(c echo.Context) error {
		return c.Render(200, "signin", page)
	})

	e.POST("/signup-validator", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		confirmPassword := c.FormValue("confirm-password")

		formData := newFormData()
		formData.Values["username"] = username
		formData.Values["password"] = password
		formData.Values["confirmPassword"] = confirmPassword
		_, existUsername := page.Data.existUsername(username)
		if len(username) < 3 {
			formData.Errors["username"] = "Username must be at least 3 characters long"
		} else if len(username) > 15 {
			formData.Errors["username"] = "Username must be at most 15 characters long"
		} else if existUsername {
			formData.Errors["username"] = "Username already exist"
		} else if page.Data.invalidUsername(username) {
			formData.Errors["username"] = "Username can only contain alphabets, numbers, underscore(_) and dash(-)"
		} else if len(password) < 6 {
			formData.Errors["password"] = "Password must be at least 6 characters long"
		} else if len(password) > 30 {
			formData.Errors["password"] = "Password must be at most 30 characters long"
		} else if confirmPassword != password {
			formData.Errors["confirmPassword"] = "Password do not match"
		}
		if len(formData.Errors) > 0 {
			return c.Render(422, "sign-up-form", formData)
		}

		user := newUser(username, password)
		page.Data.Users = append(page.Data.Users, user)
		c.Response().Header().Set("HX-Redirect", "/main")
		return nil
	})

	e.POST("/signin-validator", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		formData := newFormData()
		formData.Values["username"] = username
		formData.Values["password"] = password
		idx, existUsername := page.Data.existUsername(username)
		if !existUsername {
			formData.Errors["username"] = "Username do not exist"
			return c.Render(422, "sign-in-form", formData)
		} else if page.Data.Users[idx].Password != password {
			formData.Errors["password"] = "Password do not match"
			return c.Render(422, "sign-in-form", formData)
		}

		user := newUser(username, password)
		page.Data.Users = append(page.Data.Users, user)
		c.Response().Header().Set("HX-Redirect", "/main")
		return nil
	})

	e.GET("/main", func(c echo.Context) error {
		return c.Render(302, "main", nil)
	})

	e.Logger.Fatal(e.Start(":8000"))
}
