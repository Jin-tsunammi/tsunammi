package main

import (
	"log"
	"mm/app"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}
}

// @title						Tsunammi API
// @version					0.0.7
// @description				This is the swagger specification for Tsunammi API.
// @host						https://api-stage.tsunammi.io/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type 'Bearer' followed by a space and JWT token.
func main() {
	app.Build().Run()
}
