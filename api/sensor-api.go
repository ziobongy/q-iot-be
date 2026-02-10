package api

import (
	"net/http"
	"qiot-configuration-service/config"
	"qiot-configuration-service/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func NewSensorAPI(appConfig *config.AppConfiguration, ginEngine *gin.Engine) {
	ss := service.NewSensorService(appConfig)
	ginEngine.GET("/sensor", func(c *gin.Context) {
		getSensors(c, ss)
	})
	ginEngine.GET("/sensor/:sensorId", func(c *gin.Context) {
		getSensorById(c, ss, c.Param("sensorId"))
	})
	ginEngine.POST("/sensor", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		insertSensorConfiguration(c, ss, body)
	})
	ginEngine.PUT("/sensor/:sensorId", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		editSensorConfiguration(c, ss, c.Param("sensorId"), body)
	})
}

func getSensors(c *gin.Context, ss *service.SensorService) {
	result, err := ss.GetAllSensors()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching data from database"})
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

func getSensorById(c *gin.Context, ss *service.SensorService, sensorId string) {
	sensor, err := ss.GetSensorById(sensorId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching sensor from database"})
		return
	}
	c.IndentedJSON(http.StatusOK, sensor)
}

func insertSensorConfiguration(c *gin.Context, ss *service.SensorService, data bson.M) {
	_, errConfiguration := ss.InsertSensor(data)
	if errConfiguration != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting configuration into database"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Sensor inserted successfully"})
}
func editSensorConfiguration(c *gin.Context, ss *service.SensorService, sensorId string, data bson.M) {
	modifiedCount, errUpdate := ss.EditSensorConfiguration(
		sensorId,
		data,
	)
	if errUpdate != nil || modifiedCount == 0 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating sensor configuration in database"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Sensor updated successfully"})
}
