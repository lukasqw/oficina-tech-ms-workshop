package saga_operation

import "errors"

var (
	ErrInvalidSagaOperationID = errors.New("ID de operação de saga inválido")
	ErrInvalidSagaID          = errors.New("ID de saga inválido")
	ErrInvalidOrderID         = errors.New("ID de ordem inválido")
	ErrInvalidOperation       = errors.New("operação de saga inválida")
	ErrInvalidStatus          = errors.New("status de operação de saga inválido")
	ErrSagaOperationNotFound  = errors.New("operação de saga não encontrada")
)
