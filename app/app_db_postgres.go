package app

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (app *application) GetPgDbConnectionByName(connectionName string) (*gorm.DB, error) {
	dbConn, err := app.GetDbConnectionByName(connectionName)

	if err != nil {
		app.Logger().Error(err)
		return nil, err
	}

	if dbConn.connection == nil {
		errMsg := "connection error, connection name: " + connectionName

		app.Logger().Error(errMsg)
	}

	conn := dbConn.connection.(*gorm.DB)
	return conn, nil
}

func (app *application) createPostgresDbConnection(dbConfig dbConfig) (*gorm.DB, error) {
	dsn, _ := pq.ParseURL(dbConfig.DSN)

	app.Logger().Info("Connecting to Postgres db at ", dbConfig.Name)

	db, err := gorm.Open(postgres.Open(dsn))

	if err != nil {
		app.Logger().Panicf("Can't connect to postgres db %s error %s", dbConfig.Name, err)
		return nil, err
	}
	app.Logger().Info("Connected to Postgres db successfully at ", dbConfig.Name)

	return db, nil
}

type BaseEntityDates struct {
	CreatedAt time.Time `json:"createdAt" gorm:"default:CURRENT_TIMESTAMP;not null"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"default:CURRENT_TIMESTAMP;not null"`
}

type BaseEntity struct {
	BaseEntityDates
	ID uuid.UUID `json:"id" gorm:"primary_key; unique; type:uuid; column:id; default:uuid_generate_v4(); not null" example:"4db37f2f-a3ec-42b3-a4c8-5f6c230b0f25"`
}
