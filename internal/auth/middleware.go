package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samson/tidal-weather-clock/internal/db"
	"github.com/samson/tidal-weather-clock/internal/domain"
)

const userContextKey = "auth_user"

func SessionMiddleware(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(sessionCookie)
		if err != nil {
			c.Set(userContextKey, nil)
			c.Next()
			return
		}

		row, err := queries.GetSessionWithUser(c.Request.Context(), token)
		if err != nil {
			c.Set(userContextKey, nil)
			c.Next()
			return
		}

		c.Set(userContextKey, &domain.User{
			ID:        row.UserID,
			Email:     row.UserEmail,
			CreatedAt: row.UserCreatedAt,
		})
		c.Next()
	}
}

func RequireAuth(c *gin.Context) {
	if GetUser(c) == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}
	c.Next()
}

func GetUser(c *gin.Context) *domain.User {
	val, exists := c.Get(userContextKey)
	if !exists || val == nil {
		return nil
	}
	user, _ := val.(*domain.User)
	return user
}
