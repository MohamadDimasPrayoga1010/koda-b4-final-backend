package middleware

import (
	"fmt"
	"koda-shortlink/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware(max int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        identity := c.ClientIP() 
        endpoint := c.FullPath() 
        key := fmt.Sprintf("ratelimit:%s:%s", identity, endpoint)

        count, err := utils.RedisClient.Incr(c, key).Result()
        if err != nil {
            c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
            return
        }

        if count == 1 {
            utils.RedisClient.Expire(c, key, window)
        }

        if count > int64(max) {
            c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
            return
        }

        c.Next()
    }
}
