package repositories

import (
	"errors"
	"gorm.io/gorm"
	"payment-service/app"
	"payment-service/domain/entities"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository() *TransactionRepository {
	db, _ := app.App().GetPgDbConnectionByName("postgres")
	return &TransactionRepository{
		db: db,
	}
}

func (self TransactionRepository) BeginTx() *gorm.DB {
	tx := self.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	return tx
}

func (self TransactionRepository) CommitTx(tx *gorm.DB) *gorm.DB {
	return tx.Commit()
}

func (self TransactionRepository) RollbackTx(tx *gorm.DB) *gorm.DB {
	return tx.Rollback()
}

func (self *TransactionRepository) SaveTransaction(transaction entities.Transaction, tx *gorm.DB) (entities.Transaction, error) {
	db := self.db
	if tx != nil {
		db = tx
	}
	res := db.Save(&transaction)
	return transaction, res.Error
}

func (self *TransactionRepository) GetTransactionByTransactionId(transactionId string, tx *gorm.DB) (*entities.Transaction, error) {
	db := self.db
	if tx != nil {
		db = tx
	}
	var transaction entities.Transaction

	res := db.Model(&entities.Transaction{}).
		Where("transaction_id = ?", transactionId).
		First(&transaction)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, res.Error
	}

	return &transaction, res.Error
}

func (self *TransactionRepository) GetTransactionByChargeId(chargeId string, tx *gorm.DB) (*entities.Transaction, error) {
	db := self.db
	if tx != nil {
		db = tx
	}
	var transaction entities.Transaction

	res := db.Model(&entities.Transaction{}).
		Where("charge_id = ?", chargeId).
		First(&transaction)

	if res.Error != nil {
		return nil, res.Error
	}

	return &transaction, nil
}

func (self *TransactionRepository) GetTransactionByPaymentId(chargeId string, tx *gorm.DB) (*entities.Transaction, error) {
	db := self.db
	if tx != nil {
		db = tx
	}
	var transaction entities.Transaction

	res := db.Model(&entities.Transaction{}).
		Where("payment_id = ?", chargeId).
		First(&transaction)

	if res.Error != nil {
		return nil, res.Error
	}

	return &transaction, nil
}
