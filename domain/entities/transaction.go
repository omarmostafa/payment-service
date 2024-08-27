package entities

import "time"

type Transaction struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	TransactionType string    `gorm:"type:varchar(10);not null;check:transaction_type IN ('deposit', 'withdrawal')"`
	Amount          float64   `gorm:"type:decimal(15,2);not null"`
	Currency        string    `gorm:"type:varchar(3);not null"`
	Status          string    `gorm:"type:varchar(20);not null"`
	TransactionID   string    `gorm:"type:varchar(255) unique"`
	ChargeId        string    `gorm:"type:varchar(255)"`
	PaymentId       string    `gorm:"type:varchar(255)"`
	GatewayName     string    `gorm:"type:varchar(255)"`
	RequestPayload  *string   `gorm:"type:text"`
	ResponsePayload *string   `gorm:"type:text"`
	CallbackPayload *string   `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}
