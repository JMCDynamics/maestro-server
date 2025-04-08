package interfaces

type IUseCase[T any, R any] interface {
	Execute(props T) (R, error)
}