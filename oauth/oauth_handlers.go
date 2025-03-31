package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"myproject/db"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	bolt "go.etcd.io/bbolt"
)

type OauthHandler struct {
	KvStore   *bolt.DB
	jwtSecret string
}

var (
	ErrWrongSecret = errors.New("wrong secret")
	ErrWrongCode   = errors.New("wrong code")
)

func (h *OauthHandler) CallbackURLHandler(c *gin.Context) {
	log.Printf("RemoteAddr: %s | Params: %s", c.Request.RemoteAddr, c.Request.URL.String())
	c.Status(http.StatusNoContent)
}

func (h *OauthHandler) TokenHandler(c *gin.Context) {
	code := c.PostForm("code")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")
	// don't care about the redirect_uri for now
	if err := h.validateClientCreds(clientID, clientSecret, code); err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}

	// generate auth token, i will generate JWT, just because i don't want to store it myself.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"client_id": clientID,
		"exp":       time.Now().Add(time.Hour).Unix(),
	})

	if len(h.jwtSecret) == 0 {
		h.jwtSecret = os.Getenv("JWT_SECRET")
	}

	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.Error(fmt.Errorf("error while signing JWT: %v", err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": signed,
		"scope":        "*",
		"token_type":   "bearer",
	})
}

func (h *OauthHandler) RegistrationHandler(c *gin.Context) {
	home := c.PostForm("home")
	redirect := c.PostForm("redirect_uri")
	if len(home) == 0 {
		c.String(http.StatusBadRequest, "home not provided")
		return
	} else if len(redirect) == 0 {
		c.String(http.StatusBadRequest, "redirect_uri not provided")
		return
	}

	// check if redirect_uri and home page are on the same domain
	homeUri, err := url.Parse(home)
	if err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
	}

	redirectUri, err := url.Parse(redirect)
	if err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
	}

	if homeUri.Host != redirectUri.Host {
		c.String(http.StatusBadRequest, "hostname for home and redirect_uri doesn't match")
		return
	}

	var res db.CData
	clientID, err := uuid.NewV7()
	if err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
	}

	res.ClientID = clientID.String()
	res.ClientSecret, err = h.generateSecret(32)
	if err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
	}

	res.HomeURI = homeUri.String()
	res.RedirectURI = redirectUri.String()

	// store client_id, secret, home_url and redirect_url
	if err := db.AddClient(h.KvStore, &res); err != nil {
		c.String(http.StatusInternalServerError, "%v", err)
	}
	c.JSON(http.StatusOK, res)
}

func (h *OauthHandler) AuthHandler(c *gin.Context) {
	clientID := c.Request.URL.Query().Get("client_id")
	redirectURI := c.Request.URL.Query().Get("redirect_uri")
	// responseType := c.Request.URL.Query().Get("response_type")

	if len(clientID) == 0 || len(redirectURI) == 0 {
		c.String(http.StatusBadRequest, "")
		return
	}

	cData, err := db.GetClient(h.KvStore, clientID)
	if err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}

	if redirectURI != cData.RedirectURI {
		c.String(http.StatusBadRequest, "given redirectURI doesn't match the registered redirectURI")
		return
	}

	code, err := h.generateSecret(16)
	if err != nil {
		c.String(http.StatusInternalServerError, "")
		return
	}

	if err := db.AddCode(h.KvStore, clientID, code); err != nil {
		c.Error(fmt.Errorf("error trying to add code to DB: %v", err))
		c.String(http.StatusInternalServerError, "")
		return
	}

	c.Header("Location", redirectURI+"?code="+code)
	c.Status(http.StatusFound)
	// log.Printf("ClientID: %s, redirectURI: %s, responseType: %s", clientID, redirectURI, responseType)
}

func (h *OauthHandler) generateSecret(size int) (string, error) {
	res := make([]byte, size)
	if _, err := rand.Read(res); err != nil {
		return "", err
	}
	return hex.EncodeToString(res), nil
}

func (h *OauthHandler) validateClientCreds(clientID, clientSecret, code string) error {
	client, err := db.GetClient(h.KvStore, clientID)
	if err != nil {
		return err
	}

	if client.ClientSecret != clientSecret {
		return ErrWrongSecret
	}

	_code, err := db.GetCode(h.KvStore, client.ClientID)
	if err != nil {
		return err
	}

	if _code != code {
		return ErrWrongCode
	}
	// delete code after use
	return db.DelCode(h.KvStore, client.ClientID)
}
