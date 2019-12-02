package oauth


import (
	"github.com/deissh/osu-api-server/pkg/oauth/token"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

func empty(c echo.Context) (err error) {
	return c.JSON(200, gin.H{
		"message": "I'm alive!",
	})
}

// ApplyRoutes applies router to the gin Engine
func ApplyRoutes(r *echo.Group) {
	oauth := r.Group("/oauth")
	{
		// === Requesting Tokens ===
		oauth.POST("/token", token.CreateTokenHandler) // https://laravel.com/docs/5.7/passport#refreshing-tokens

		// === Managing Tokens ===
		oauth.GET("/scopes", empty)

		// === Managing Clients ===
		oauth.GET("/clients", empty)
		oauth.POST("/clients", empty)
		oauth.PUT("/clients/:client", empty)
	}
}