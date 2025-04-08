package interfaces

import "github.com/JMCDynamics/maestro-server/internal/dtos"

type IVpnGateway interface {
	GenerateNewPeer(name string) (dtos.ResponseNewPeer, error)
	Run() error
}
