package main

import (
	"log"

	"myproject/db"
	"myproject/oauth"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/mattn/go-sqlite3"
)

// var responseType = []string{
// 	"code",
// 	"token",
// 	"id_token",
// 	"code token",
// 	"code id_token",
// 	"token id_token",
// 	"code token id_token",
// }

func init() {
	if err := godotenv.Load("./.env.local"); err != nil {
		log.Fatal(err)
	}
}

// oauth server
// server needs to know the client (client needs to register with the auth server before-hand)
// client oauth steps
//  1. client request to /oauth/authorize?client_id=ID&redirect_uri=URI&scope=scope&response_type=code&state=state
//  2. resource owner authorizes it, or logs-in and authorizes, then oauth server sends a response
//     with a code to the redirect_uri
//  3. client sends a request with the code to the resource server to get an access token
//  4. client uses the access_token for further requests
func main() {
	_db, err := db.InitDB("./.sqlite3.db")
	if err != nil {
		log.Fatal(err)
	}

	_kv, err := db.InitKvStore("./.bolt.db")
	if err != nil {
		_db.Close()
		log.Fatal(err)
	}

	defer _kv.Close()
	defer _db.Close()

	handler := oauth.Handler{
		Db: _db,
	}

	oauthHandler := oauth.OauthHandler{
		KvStore: _kv,
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

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
	router.Run(":80")
}
