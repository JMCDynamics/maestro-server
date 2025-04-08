package usecases

import (
	"context"
	"fmt"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

type responseDefaultUser struct {
	AlreadyExists bool   `json:"alreadyExists"`
	VpnConfig     string `json:"vpnConfig"`
}

type CreateDefaultUserUseCase struct {
	databaseGateway interfaces.IDatabaseGateway
	vpnGateway      interfaces.IVpnGateway
}

func NewCreateDefaultUserUseCase(
	databaseGateway interfaces.IDatabaseGateway,
	vpnGateway interfaces.IVpnGateway,
) interfaces.IUseCase[dtos.CreateUserDTO, responseDefaultUser] {
	return &CreateDefaultUserUseCase{
		databaseGateway: databaseGateway,
		vpnGateway:      vpnGateway,
	}
}

func (u *CreateDefaultUserUseCase) Execute(data dtos.CreateUserDTO) (responseDefaultUser, error) {
	id := ulid.Make().String()

	sql := "SELECT id FROM users WHERE username = $1 LIMIT 1"
	result, err := u.databaseGateway.Query(context.Background(), sql, data.Username)
	if err != nil {
		return responseDefaultUser{}, err
	}
	if result.Next() {
		return responseDefaultUser{
			AlreadyExists: true,
		}, nil
	}

	sql = "DELETE FROM users"
	if err := u.databaseGateway.Exec(context.Background(), sql); err != nil {
		return responseDefaultUser{}, fmt.Errorf("unable to ensure only one default user: %v", err)
	}

	passwordHashed, err := hashPassword(data.Password)
	if err != nil {
		return responseDefaultUser{}, err
	}
	sql = "INSERT INTO users (id, username, password) VALUES($1,$2,$3)"
	if err := u.databaseGateway.Exec(context.Background(), sql, id, data.Username, passwordHashed); err != nil {
		return responseDefaultUser{}, fmt.Errorf("unable to create default user: %v", err)
	}

	peer, _ := u.vpnGateway.GenerateNewPeer(data.Username)

	return responseDefaultUser{
		AlreadyExists: false,
		VpnConfig:     peer.ConfigOutput,
	}, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
