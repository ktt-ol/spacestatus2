package web

import (
	"time"
	"github.com/gin-gonic/gin"
	"fmt"
)

func SimpleNoTimeLogging() gin.HandlerFunc {

	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		fmt.Fprintln(gin.DefaultWriter,
			c.Writer.Status(),
			"|",
			c.ClientIP(),
			"|",
			latency,
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
		)
	}
}
