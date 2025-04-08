package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
)

type FindNodesUseCase struct {
	databaseGateway interfaces.IDatabaseGateway
	cacheGateway    interfaces.ICacheGateway
}

func NewFindNodesUseCase(
	databaseGateway interfaces.IDatabaseGateway,
	cacheGateway interfaces.ICacheGateway,
) interfaces.IUseCase[any, []dtos.Node] {
	return &FindNodesUseCase{
		databaseGateway: databaseGateway,
		cacheGateway:    cacheGateway,
	}
}

func (u *FindNodesUseCase) Execute(_ any) ([]dtos.Node, error) {
	sql := "SELECT id, name, operating_system, vpn_address FROM nodes"
	resultSet, err := u.databaseGateway.Query(context.Background(), sql)
	if err != nil {
		return []dtos.Node{}, errors.New("unable to find nodes")
	}
	defer resultSet.Close()

	nodes := []dtos.Node{}
	for resultSet.Next() {
		var node dtos.Node
		if err := resultSet.Scan(&node.Id, &node.Name, &node.OperatingSystem, &node.VpnAddress); err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		status, err := u.cacheGateway.Get(context.Background(), node.Id)
		if err != nil {
			status = dtos.DOWN
		}

		node.Status = status
		nodes = append(nodes, node)
	}

	return nodes, nil
}
