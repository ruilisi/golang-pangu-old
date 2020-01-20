package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"qiyetalk-server-go/db"
	"qiyetalk-server-go/models"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var identityKey = "email"
var jwtKey = []byte("my_secret_key")

// Credentials ...
type Credentials struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(identityKey)
	fmt.Println(claims)
	c.JSON(200, gin.H{
		"userID": claims[identityKey],
		"email":  user.(*models.User).Email,
		"text":   "Hello World.",
	})
}

// Signup ...
func Signup(c *gin.Context) {
	creds := &Credentials{}
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_db := db.GetDB()
	encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	err := _db.Insert(&models.User{
		Email:             creds.Email,
		EncryptedPassword: string(encryptedPassword),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(200)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "qiyetalk",
		Key:         jwtKey,
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					identityKey: v.Email,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.User{
				Email: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			creds := &Credentials{}
			if err := c.ShouldBind(&creds); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			user := models.FindByEmail(creds.Email)
			if user != nil {
				err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(creds.Password))
				fmt.Println(err)
				if err == nil {
					return &models.User{
						Email: creds.Email,
					}, nil
				} else {
					return nil, err
				}
			}
			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*models.User); ok && v.Email == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	r.POST("/users/sign_in", authMiddleware.LoginHandler)
	r.POST("/users/sign_up", Signup)

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	auth := r.Group("/auth")
	// Refresh time can be longer than token timeout
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/hello", helloHandler)
	}

	r.Run(":" + port)
}
