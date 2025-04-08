package middlewares

import (
	"net/http"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type authMiddleware struct {
	maestroSecretKey string
}

func NewAuthMiddleware(maestroSecretKey string) authMiddleware {
	return authMiddleware{
		maestroSecretKey: maestroSecretKey,
	}
}

// TODO: Quero alterar o metodo de cookie para Bearer token no header

func (a *authMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearerToken := c.GetHeader("Authorization")

		if bearerToken == "" {
			response := dtos.NewDefaultResponse("token not provided", nil)
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		if len(bearerToken) < 7 || bearerToken[:7] != "Bearer " {
			response := dtos.NewDefaultResponse("invalid token", nil)
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		tokenString := bearerToken[7:]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(a.maestroSecretKey), nil
		})

		if err != nil || !token.Valid {
			response := dtos.NewDefaultResponse("invalid or expired token", nil)
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("userId", claims["userId"])
		}

		c.Next()
	}
}
