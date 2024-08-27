package main

import (
	"payment-service/app"
	"payment-service/domain/entities"
	"payment-service/routes"
)

func main() {
	app := app.App()

	db, _ := app.GetPgDbConnectionByName("postgres")
	db.AutoMigrate(&entities.Transaction{})

	defer app.Clean()
	app.SetRoutes(routes.GetRoutes())
	app.StartServer()
}
