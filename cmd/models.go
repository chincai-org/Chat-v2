package main

import (
	"net/http"
	"fmt"
	"sync"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo-contrib/session"
	"github.com/gorilla/sessions"
)

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
	users map[string]User
}{users: make(map[string]User)}

// sessionStore holds the cookie-based session store.
// NOTE: In production, the key must remain constant across restarts.
var sessionStore = sessions.NewCookieStore([]byte("this-key-changes-each-reload" + strconv.FormatInt(time.Now().Unix(), 10)))

// getPassword retrieves the password for a user by index.
func getPassword(username string) (string, error) {
    userStore.RLock()
    defer userStore.RUnlock()
    
    user, exists := userStore.users[username]
    if !exists {
        return "", fmt.Errorf("user %q not found", username)
    }
    return user.Password, nil
}

// existUsername checks if a username already exists.
func existUsername(username string) bool {
	userStore.RLock()
	defer userStore.RUnlock()

	_, exists := userStore.users[username]
    return exists
}

// invalidUsername determines whether a username contains invalid characters.
func invalidUsername(username string) bool {
	for _, char := range username {
		if (char < 'A' || char > 'Z') &&
			(char < 'a' || char > 'z') &&
			(char < '0' || char > '9') &&
			char != '_' && char != '-' {
			return true
		}
	}
	return false
}

// save adds the user to the user store.
func (u *User) save() error {
	userStore.Lock()
	defer userStore.Unlock()

	if _, exists := userStore.users[u.Username]; exists {
        return fmt.Errorf("username exists")
    }

	userStore.users[u.Username] = *u
	return nil
}

// login handles user authentication and session setup.
func (u *User) login(c echo.Context) error {
	// Set up session
	sess, err := session.Get("session", c)
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

	return c.Render(http.StatusOK, "index", map[string]string{"Username": u.Username})
}
