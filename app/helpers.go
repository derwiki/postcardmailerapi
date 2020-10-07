package helpers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func SetCorsHeaders(c *gin.Context) {
	var origin = c.GetHeader("Origin")
	fmt.Println("origin", origin)
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Content-Type", `application/json`)
	c.Header("Accept", `application/json`)
}
