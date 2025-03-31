package oauth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"myproject/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Db *sql.DB
}

// signed in root page
func (h *Handler) Root(c *gin.Context) {
	session, err := c.Cookie("session")
	if err != nil || len(session) == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user, err := db.GetUserBySession(h.Db, session)
	if err != nil {
		log.Print(err)
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"username": user.Username,
	})
}

func (h *Handler) GetLoginHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "test",
	})
}

func (h *Handler) Signup(c *gin.Context) {
	username := c.PostForm("Username")
	password := c.PostForm("Password")
	email := c.PostForm("email") // ideally validation is needed because a user can use the terminal to send a request

	if len(username) < 4 || len(password) < 4 {
		c.Status(http.StatusBadRequest)
		return
	}

	if res, err := h.validateUserInfo(username, email); err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
		return
	} else if len(res) > 0 {
		c.String(http.StatusBadRequest, "%s", res)
		return
	}

	session, err := h.generateSession()
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}

	user := db.User{
		Username: username,
		Password: password,
		Email:    email,
		Session:  session,
	}

	if err := db.CreateUser(h.Db, &user); err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
		return
	}

	c.SetCookie("session", user.Session, 3600, "/", "", false, true)
	c.HTML(http.StatusOK, "created.html", nil)
}

func (h *Handler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	// log.Printf("Username: %s and Password: %s\n", Username, Password)
	// this is where a database comes in
	if err := h.validateLogin(username, password); err != nil {
		c.String(http.StatusUnauthorized, "%v", err)
		return
	}

	session, _ := h.generateSession()
	log.Print(session)
	if err := db.SetSession(h.Db, username, session); err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}

	c.SetCookie("session", session, 3600, "/", "", false, true)
	c.HTML(200, "index.html", gin.H{
		"username": username,
	})
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie("session")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		// validate session
		if _, err := db.GetUserBySession(h.Db, session); err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *Handler) generateSession() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *Handler) validateLogin(Username, Password string) error {
	user, err := db.GetUser(h.Db, Username)
	if err != nil {
		return err
	}

	if user == nil || user.Password != Password {
		return errors.New("wrong Username/Password")
	}
	return nil
}

func (h *Handler) validateUserInfo(Username, email string) (string, error) {
	res, err := db.GetUser(h.Db, Username)
	if err != nil {
		return "", err
	}

	if res == nil { // what i want
		if ok, err := db.EmailExists(h.Db, email); err != nil {
			return "", err
		} else if ok {
			return "email exists", nil
		}
		return "", nil
	}
	return "user exists", nil
}
