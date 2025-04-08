package services

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

type NodeWebSocketService struct {
	m sync.Mutex

	connections map[string]*websocket.Conn
}

func NewNodeWebSocketService() NodeWebSocketService {
	return NodeWebSocketService{
		connections: make(map[string]*websocket.Conn),
	}
}

func (s *NodeWebSocketService) AddConnection(conn *websocket.Conn, nodeId string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.connections[nodeId]; ok {
		return errors.New("unable to establish a connection with maestro-server")
	}

	s.connections[nodeId] = conn
	return nil
}

func (s *NodeWebSocketService) RemoveConnection(id string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.connections, id)
}
