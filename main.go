package main

import (
	"qiot-configuration-service/api"
	"qiot-configuration-service/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	appConfiguration := config.NewAppConfiguration()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	api.NewConfigurationAPI(appConfiguration, router)
	api.NewSensorAPI(appConfiguration, router)
	api.NewExperimentAPI(appConfiguration, router)

	router.Run(":8080")
}
