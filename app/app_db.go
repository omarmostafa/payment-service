package app

import (
	"errors"
	_ "github.com/lib/pq"
)

func (app *application) GetDbConnectionByName(connName string) (dbConnection, error) {

	errMsg := "connection is not defined, connection name: " + connName

	if val, ok := app.dbConnections[connName]; ok {
		return val, nil
	}

	app.Logger().Error(errMsg)

	return dbConnection{}, errors.New(errMsg)
}

func (app *application) setupDbConnections() *application {

	app.Logger().Debug("setup DB Connections")

	dbConnections := make(map[string]dbConnection)

	for _, dbConfig := range app.db {
		dbConnection := dbConnection{}
		dbConnection.config = dbConfig
		dbConnection.name = dbConfig.Name
		conn, err := app.createDbConnection(dbConfig)
		if err != nil {
			app.Logger().Panic(err)
		} else {
			dbConnection.connection = conn
			dbConnections[dbConfig.Name] = dbConnection
		}

	}
	app.dbConnections = dbConnections
	return app
}

func (app *application) Clean() *application {

	app.Logger().Debug("clean before shutdown")

	app.CloseConnections()

	return app
}

func (app *application) CloseConnections() *application {

	app.Logger().Debugf("CloseDbConnections")

	dbConnections := make(map[string]dbConnection)
	app.dbConnections = dbConnections
	return app
}

func (app *application) createDbConnection(dbConfig dbConfig) (interface{}, error) {

	switch dbConfig.Driver {

	case DBDriverPostgres:
		return app.createPostgresDbConnection(dbConfig)
	}

	return nil, errors.New("Can't handle connection for driver [" + dbConfig.Driver + "] of connection [" + dbConfig.Name + "]")
}
