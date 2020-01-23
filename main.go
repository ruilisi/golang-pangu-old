package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"qiyetalk-server-go/db"
	"qiyetalk-server-go/models"
	"qiyetalk-server-go/utils"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(models.IdentityKey)
	fmt.Println(claims)
	c.JSON(200, gin.H{
		"userID": claims[models.IdentityKey],
		"email":  user.(*models.User).Email,
		"text":   "Hello World.",
	})
}

// Signup ...
func Signup(c *gin.Context) {
	type Data struct {
		User models.Credentials `json:"user" binding:"required"`
	}

	data := &Data{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	creds := data.User

	if creds.Password != creds.PasswordConfirmation {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your password and confirmation password do not match"})
		return
	}

	_db := db.GetDB()

	user := models.FindByEmail(creds.Email)
	if user != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your account has been registered"})
		return
	}
	encryptedPassword, _ := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	err := _db.Insert(&models.User{
		Email:             creds.Email,
		EncryptedPassword: string(encryptedPassword),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authMiddleware, err := utils.JwtMiddleWare()

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	r.POST("/users/sign_in", authMiddleware.LoginHandler)
	r.POST("/users", Signup)

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"wechat_app_id": os.Getenv("WECHAT_APP_ID")})
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
