package usecases

import (
	"context"
	"fmt"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
)

type UpdateNodeUseCase struct {
	databaseGateway interfaces.IDatabaseGateway
	cacheGateway    interfaces.ICacheGateway
}

func NewUpdateNodeUseCase(
	databaseGateway interfaces.IDatabaseGateway,
	cacheGateway interfaces.ICacheGateway,
) interfaces.IUseCase[dtos.UpdateNodeDTO, dtos.Node] {
	return &UpdateNodeUseCase{
		databaseGateway: databaseGateway,
		cacheGateway:    cacheGateway,
	}
}

func (u *UpdateNodeUseCase) Execute(data dtos.UpdateNodeDTO) (dtos.Node, error) {
	var node dtos.Node
	sqlFind := "SELECT id, name, operating_system FROM nodes WHERE id = $1"
	resultSet, err := u.databaseGateway.Query(context.Background(), sqlFind, data.Id)
	if err != nil {
		return dtos.Node{}, fmt.Errorf("unable to find a node: %v", err)
	}

	if !resultSet.Next() {
		return dtos.Node{}, ErrNodeNotFound
	}

	if err := resultSet.Scan(&node.Id, &node.Name, &node.OperatingSystem); err != nil {
		return dtos.Node{}, fmt.Errorf("unable to find a node: %v", err)
	}

	sql := "UPDATE nodes SET name = $1, operating_system = $2 WHERE id = $3"
	if err := u.databaseGateway.Exec(context.Background(), sql, data.Name, data.OperatingSystem, data.Id); err != nil {
		return dtos.Node{}, fmt.Errorf("unable to create a node: %v", err)
	}

	var status dtos.TypeNodeStatus = dtos.DOWN
	value, _ := u.cacheGateway.Get(context.Background(), node.Id)
	if value != "" {
		status = dtos.TypeNodeStatus(value)
	}

	return dtos.Node{
		Id:              data.Id,
		Name:            data.Name,
		OperatingSystem: data.OperatingSystem,
		VpnAddress:      node.VpnAddress,
		Status:          status,
	}, nil
}
