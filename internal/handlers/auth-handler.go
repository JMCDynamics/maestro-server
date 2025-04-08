package handlers

import (
	"net/http"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	usecases "github.com/JMCDynamics/maestro-server/internal/use-cases"
	"github.com/gin-gonic/gin"
)

type authHandler struct {
	authenticateUserUseCase interfaces.IUseCase[dtos.AuthUserDTO, string]
}

func NewAuthHandler(
	authenticateUserUseCase interfaces.IUseCase[dtos.AuthUserDTO, string],
) authHandler {
	return authHandler{
		authenticateUserUseCase: authenticateUserUseCase,
	}
}

func (h *authHandler) HandleAuth(c *gin.Context) {
	var body dtos.AuthUserDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		response := dtos.NewDefaultResponse("invalid request body", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	token, err := h.authenticateUserUseCase.Execute(body)
	if err != nil {
		response := dtos.NewDefaultResponse("unable to authenticate user", err.Error())
		status := http.StatusInternalServerError
		if err == usecases.ErrInvalidCredentials {
			status = http.StatusUnauthorized
		}
		c.JSON(status, response)
		return
	}

	response := dtos.NewDefaultResponse("action exectued with success", token)
	c.JSON(http.StatusOK, response)
}

func (h *authHandler) HandleLogout(c *gin.Context) {
	response := dtos.NewDefaultResponse("logged out successfully", nil)
	c.JSON(http.StatusOK, response)
}
