package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func FetchHealth(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"status":  "success",
			"message": "API is healthy",
		},
	)
	fmt.Println("Health check done")
}
