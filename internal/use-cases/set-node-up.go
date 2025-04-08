package usecases

import (
	"context"
	"time"

	"github.com/JMCDynamics/maestro-server/internal/dtos"
	"github.com/JMCDynamics/maestro-server/internal/interfaces"
)

type SetNodeUpUseCase struct {
	cacheGateway interfaces.ICacheGateway
}

func NewSetNodeUpUseCase(
	cacheGateway interfaces.ICacheGateway,
) interfaces.IUseCase[string, any] {
	return &SetNodeUpUseCase{
		cacheGateway: cacheGateway,
	}
}

func (u *SetNodeUpUseCase) Execute(id string) (any, error) {
	if err := u.cacheGateway.Set(context.Background(), id, dtos.UP, 5*time.Second); err != nil {
		return nil, err
	}

	return nil, nil
}
