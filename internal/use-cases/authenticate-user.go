package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/JMCDynamics/maestro-server/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticateUserUseCase struct {
	databaseGateway  interfaces.IDatabaseGateway
	maestroSecretKey string
}

var (
	ErrInvalidCredentials error = errors.New("invalid username or password")
)

func NewAuthenticateUserUseCase(
	databaseGateway interfaces.IDatabaseGateway,
	maestroSecretKey string,
) interfaces.IUseCase[dtos.AuthUserDTO, string] {
	return &AuthenticateUserUseCase{
		databaseGateway:  databaseGateway,
		maestroSecretKey: maestroSecretKey,
	}
}

func (u *AuthenticateUserUseCase) Execute(data dtos.AuthUserDTO) (string, error) {
	sql := "SELECT id, password FROM users WHERE username = $1 LIMIT 1"
	result, err := u.databaseGateway.Query(context.Background(), sql, data.Username)
	if err != nil {
		return "", err
	}
	if !result.Next() {
		return "", ErrInvalidCredentials
	}

	var id string
	var hashedPassword string
	if err := result.Scan(&id, &hashedPassword); err != nil {
		return "", err
	}

	isValid := checkPasswordHash(data.Password, hashedPassword)
	if !isValid {
		return "", ErrInvalidCredentials
	}

	jwtToken, err := utils.GenerateJWT(id, u.maestroSecretKey)
	if err != nil {
		return "", fmt.Errorf("unable to generate token")
	}

	return jwtToken, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
