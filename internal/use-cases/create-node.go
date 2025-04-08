package usecases

import (
	"context"
	"fmt"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/oklog/ulid/v2"
)

type CreateNode struct {
	databaseGateway interfaces.IDatabaseGateway
	vpnGateway      interfaces.IVpnGateway
}

func NewCreateNode(
	databaseGateway interfaces.IDatabaseGateway,
	vpnGateway interfaces.IVpnGateway,
) interfaces.IUseCase[dtos.CreateNodeDTO, dtos.Node] {
	return &CreateNode{
		databaseGateway: databaseGateway,
		vpnGateway:      vpnGateway,
	}
}

func (u *CreateNode) Execute(data dtos.CreateNodeDTO) (dtos.Node, error) {
	id := ulid.Make().String()

	config, err := u.vpnGateway.GenerateNewPeer(id)
	if err != nil {
		return dtos.Node{}, err
	}

	sql := "INSERT INTO nodes (id, name, vpn_address, operating_system) VALUES($1,$2,$3,$4)"
	if err := u.databaseGateway.Exec(context.Background(), sql, id, data.Name, config.VpnAddress, data.OperatingSystem); err != nil {
		return dtos.Node{}, fmt.Errorf("unable to create a node: %v", err)
	}

	return dtos.Node{
		Id:              id,
		Name:            data.Name,
		OperatingSystem: data.OperatingSystem,
		VpnAddress:      config.VpnAddress,
		Status:          dtos.DOWN,
	}, nil
}
