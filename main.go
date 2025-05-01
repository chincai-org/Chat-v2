package main

import (
	"time"
	"html/template"
	"io"
	"strconv"
	"net/http"
	"sync"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/echo-contrib/session"
	"github.com/gorilla/sessions"
)

const StatusUnprocessableEntity = 422

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
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

var userStore = struct {
	sync.RWMutex
	users []User
}{users: []User{}}

func getPassword(index int) (string, error) {
	userStore.RLock()
	defer userStore.RUnlock()

	if index < 0 || index >= len(userStore.users) {
		return "", fmt.Errorf("index out of bounds: %d", index)
	}

	return userStore.users[index].Password, nil
}

func existUsername(username string) (int, bool) {
	userStore.RLock() // Acquire read lock
	defer userStore.RUnlock()

	for i, user := range userStore.users {
		if user.Username == username {
			return i, true
		}
	}
	return 0, false
}

func invalidUsername(username string) bool {
	for _, char := range username {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && char != '_' && char != '-' {
			return true
		}
	}
	return false
}

func (u *User) login(c echo.Context) error {
	userStore.Lock()
	userStore.users = append(userStore.users, *u)
	userStore.Unlock()

	// Set up session
	sess, err := session.Get("session", c)
	// Corrupted session from session key change
	// if err != nil || sess == nil {
		// return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read session")
	// }
	// Debug fix
	if err != nil || sess == nil {
		sess = sessions.NewSession(sessionStore, "session")
	}

	sess.Options = &sessions.Options {
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
	}
	sess.Values["username"] = u.Username
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save session")
	}

	c.Response().Header().Set("HX-Redirect", "/")
	return c.NoContent(http.StatusOK)
}

func checkLoginSession(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err == nil && sess != nil {
			username, ok := sess.Values["username"].(string)
			if ok && username != "" {
				return c.Redirect(http.StatusSeeOther, "/")
			}
		}
		return next(c)
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

// Use unchanging key in release
var sessionStore = sessions.NewCookieStore([]byte("this-key-change-each-reload"+strconv.FormatInt(time.Now().Unix(), 10)))

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(sessionStore))
	// Force reload if not htmx request
	e.Use(func (next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Header.Get("HX-Request") == "" {
				c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Response().Header().Set("Pragma", "no-cache")
				c.Response().Header().Set("Expires", "0")
			}
			return next(c)
		}
	})

	e.Renderer = newTemplate()
	e.Static("/static/images", "images")
	e.Static("/static/css", "css")

	e.GET("/", func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil || sess == nil {
			return c.Render(http.StatusOK, "index", nil)
		}
		username, ok := sess.Values["username"].(string)
		if !ok || username == "" {
			return c.Render(http.StatusOK, "index", nil)
		}
		return c.Render(http.StatusOK, "main", map[string]string{"Username": username})
	})

	e.GET("/signup", func(c echo.Context) error {
		return c.Render(http.StatusOK, "signup", nil)
	}, checkLoginSession)

	e.GET("/signin", func(c echo.Context) error {
		return c.Render(http.StatusOK, "signin", nil)
	}, checkLoginSession)

	e.POST("/signup-validator", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		confirmPassword := c.FormValue("confirm-password")

		formData := newFormData()
		formData.Values["username"] = username
		formData.Values["password"] = password
		formData.Values["confirmPassword"] = confirmPassword
		_, existUsername := existUsername(username)
		if len(username) < 3 {
			formData.Errors["username"] = "Username must be at least 3 characters long"
		} else if len(username) > 15 {
			formData.Errors["username"] = "Username must be at most 15 characters long"
		} else if existUsername {
			formData.Errors["username"] = "Username already exist"
		} else if invalidUsername(username) {
			formData.Errors["username"] = "Username can only contain alphabets, numbers, underscore(_) and dash(-)"
		} else if len(password) < 6 {
			formData.Errors["password"] = "Password must be at least 6 characters long"
		} else if len(password) > 30 {
			formData.Errors["password"] = "Password must be at most 30 characters long"
		} else if confirmPassword != password {
			formData.Errors["confirmPassword"] = "Password do not match"
		}
		if len(formData.Errors) > 0 {
			return c.Render(StatusUnprocessableEntity, "sign-up-form", formData)
		}

		user := newUser(username, password)
		return user.login(c)
	}, checkLoginSession)

	e.POST("/signin-validator", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		formData := newFormData()
		formData.Values["username"] = username
		formData.Values["password"] = password
		index, existUsername := existUsername(username)
		if !existUsername {
			formData.Errors["username"] = "Username do not exist"
			return c.Render(StatusUnprocessableEntity, "sign-in-form", formData)
		} 
		readPassword, err := getPassword(index)
		if err != nil {
			fmt.Println("Error retrieving password:", err)
			formData.Errors[""] = "An unexpected error occurred. Please try again."
			return c.Render(http.StatusInternalServerError, "sign-in-form", formData)
		} 
		if (readPassword != password) {
			formData.Errors["password"] = "Password do not match"
			return c.Render(StatusUnprocessableEntity, "sign-in-form", formData)
		}

		user := newUser(username, password)
		return user.login(c)
	}, checkLoginSession)

	e.Logger.Fatal(e.Start(":8000"))
}
