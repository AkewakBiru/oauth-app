package oauth_test

import (
	"bytes"
	"myproject/db"
	"myproject/oauth"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var router *gin.Engine

const dbFile = "test.db"

func setup() {
	gin.SetMode(gin.TestMode)
	router = gin.New()
	router.LoadHTMLGlob("../templates/*")
	handler := oauth.Handler{}
	oauthHandler := oauth.OauthHandler{}
	router.GET("/", handler.Root)
	router.GET("/index", handler.Root)
	router.GET("/signup", func(c *gin.Context) {
		c.HTML(200, "signup.html", nil)
	})
	router.POST("/signup", handler.Signup)
	router.POST("/login", handler.Login)
	router.GET("/login", handler.GetLoginHandler)
	router.POST("/register", oauthHandler.RegistrationHandler) // client registers to get an id and secret

	auth := router.Group("/")
	auth.Use(handler.AuthMiddleware())
	{
		auth.GET("/oauth/authorize", oauthHandler.AuthHandler)
		auth.POST("/oauth/token", oauthHandler.TokenHandler)
		auth.GET("/oauth2/callback", oauthHandler.CallbackURLHandler) // transfer to client when done testing
	}
}

func TestMain(m *testing.M) {
	setup()
	m.Run()
}

func TestLoginGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.EqualValues(t, http.StatusOK, w.Code)
}

func TestRootRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.EqualValues(t, http.StatusFound, w.Code)
}

func TestLoginFail(t *testing.T) {
	// there needs to be a setup database
	db, err := db.InitDB(dbFile)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer db.Close()
	defer os.Remove(dbFile)

	data := "username=test&password=test"
	body := bytes.NewReader([]byte(data))
	req := httptest.NewRequest(http.MethodPost, "/login", body)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.EqualValues(t, http.StatusUnauthorized, w.Code)
}

func TestLoginSuccess(t *testing.T) {

}
