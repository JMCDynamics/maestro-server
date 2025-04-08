package usecases

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
)

type FindNodeUseCase struct {
	databaseGateway interfaces.IDatabaseGateway
	cacheGateway    interfaces.ICacheGateway
}

var (
	ErrNodeNotFound error = errors.New("failed to search for node")
)

func NewFindNodeUseCase(
	databaseGateway interfaces.IDatabaseGateway,
	cacheGateway interfaces.ICacheGateway,
) interfaces.IUseCase[string, dtos.Node] {
	return &FindNodeUseCase{
		databaseGateway: databaseGateway,
		cacheGateway:    cacheGateway,
	}
}

func (u *FindNodeUseCase) Execute(id string) (dtos.Node, error) {
	sql := "SELECT id, name, operating_system, vpn_address FROM nodes WHERE id = $1"
	resultSet, err := u.databaseGateway.Query(context.Background(), sql, id)
	if err != nil {
		return dtos.Node{}, errors.New("unable to find node")
	}
	defer resultSet.Close()

	if !resultSet.Next() {
		return dtos.Node{}, ErrNodeNotFound
	}

	var node dtos.Node
	if err := resultSet.Scan(&node.Id, &node.Name, &node.OperatingSystem, &node.VpnAddress); err != nil {
		return dtos.Node{}, fmt.Errorf("failed to scan node: %w", err)
	}

	status, err := u.cacheGateway.Get(context.Background(), node.Id)
	if err != nil {
		status = dtos.DOWN
	}

	configPath := fmt.Sprintf("/config/peer_%s/peer_%s.conf", node.Id, node.Id)

	contentBytes, _ := os.ReadFile(configPath)

	contentString := string(contentBytes)
	if contentString != "" {
		fmt.Println(base64.StdEncoding.EncodeToString(
			[]byte(strings.ReplaceAll(contentString, "\"", "")),
		))
		node.VpnConfig = base64.StdEncoding.EncodeToString(
			[]byte(strings.ReplaceAll(contentString, "\"", "")),
		)
	}

	node.Status = status
	return node, nil
}
