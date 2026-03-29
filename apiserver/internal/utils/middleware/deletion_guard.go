package middleware

import (
	"net/http"
	"strings"

	authUtils "dkhalife.com/tasks/core/internal/utils/auth"
	"github.com/gin-gonic/gin"
)

var deletionExemptPaths = map[string]struct{}{
	"/api/v1/users/deletion": {},
}

// DeletionGuardMiddleware blocks write operations for accounts that are pending deletion.
// Read-only methods (GET, HEAD, OPTIONS) are always permitted. The deletion management
// endpoints themselves are also exempt so users can cancel.
func DeletionGuardMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		// Normalize: strip trailing slash for comparison
		path = strings.TrimRight(path, "/")
		if _, exempt := deletionExemptPaths[path]; exempt {
			c.Next()
			return
		}

		identity := authUtils.CurrentIdentity(c)
		if identity != nil && identity.PendingDeletion {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Account is pending deletion",
			})
			return
		}

		c.Next()
	}
}
