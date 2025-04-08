package services

import "github.com/JMCDynamics/maestro-server/internal/dtos"

type NodeStatusService struct {
	c chan dtos.NodeStatus
}

func NewNodeStatusService() *NodeStatusService {
	return &NodeStatusService{
		c: make(chan dtos.NodeStatus),
	}
}

func (n *NodeStatusService) ListenStatus() <-chan dtos.NodeStatus {
	return n.c
}

func (n *NodeStatusService) SetStatus(status dtos.NodeStatus) {
	n.c <- status
}
