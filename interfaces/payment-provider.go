package interfaces

import (
	"payment-service/domain/entities"
	"payment-service/domain/types"
)

type IPaymentProvider interface {
	Charge(params types.DepositParams, transaction entities.Transaction) (entities.Transaction, error)
	Withdraw(params types.WithdrawParams, transaction entities.Transaction) (entities.Transaction, error)
}
