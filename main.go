package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

var (
	ProjectID   = os.Getenv("SANITY_PROJECT_ID")
	Version     = os.Getenv("SANITY_VERSION")
	Token       = os.Getenv("SANITY_TOKEN")
	DeeplAPIURL = "https://api-free.deepl.com/v2/translate"
	BaseAPIURL  = "https://%s.api.sanity.io/%s/data"
)

func main() {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(SetUpCORS())

	router.POST("/sanity_translate_document", SanityTranslateDocument)
	router.POST("/sanity_translate_field", SanityTranslateField)

	router.GET("/health", FetchHealth)

	fmt.Println("Starting Sanity Translation Service")
	fmt.Println("")

	router.Run(":8080")

}

func SetUpCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")
		c.Next()
	}
}
