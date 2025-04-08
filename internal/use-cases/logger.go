package usecases

import (
	"time"

	"github.com/JMCDynamics/maestro-server/internal/interfaces"
	"github.com/rs/zerolog/log"
)

type LoggerUseCase[T any, R any] struct {
	actor interfaces.IUseCase[T, R]
}

func NewLoggerUseCase[T any, R any](actor interfaces.IUseCase[T, R]) *LoggerUseCase[T, R] {
	return &LoggerUseCase[T, R]{actor: actor}
}

func (u *LoggerUseCase[T, R]) Execute(props T) (R, error) {
	start := time.Now()

	log.Debug().
		Str("event", "use_case_execution").
		Str("use_case", "LoggerUseCase").
		Interface("input", props).
		Time("timestamp", start).
		Msg("executing use case")

	result, err := u.actor.Execute(props)

	duration := time.Since(start)

	if err != nil {
		log.Error().
			Str("event", "use_case_failed").
			Str("use_case", "LoggerUseCase").
			Err(err).
			Time("timestamp", time.Now()).
			Dur("execution_time", duration).
			Msg("use case execution failed")
		return result, err
	}

	log.Debug().
		Str("event", "use_case_success").
		Str("use_case", "LoggerUseCase").
		Interface("output", result).
		Time("timestamp", time.Now()).
		Dur("execution_time", duration).
		Msg("use case executed successfully")

	return result, nil
}
