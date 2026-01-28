package api

import (
	"log"
	"net/http"
	"qiot-configuration-service/config"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewSensorAPI(appConfig *config.AppConfiguration, ginEngine *gin.Engine) {
	ginEngine.GET("/sensor", func(c *gin.Context) {
		getSensors(c, appConfig)
	})
	ginEngine.GET("/sensor/:sensorId", func(c *gin.Context) {
		getSensorById(c, appConfig, c.Param("sensorId"))
	})
	ginEngine.POST("/sensor", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		insertSensorConfiguration(c, appConfig, body)
	})
	ginEngine.PUT("/sensor/:sensorId", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		editSensorConfiguration(c, appConfig, c.Param("sensorId"), body)
	})
}

func getSensors(c *gin.Context, appConfig *config.AppConfiguration) {
	result, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{})
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching data from database"})
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

func getSensorById(c *gin.Context, appConfig *config.AppConfiguration, sensorId string) {
	oid, err2 := primitive.ObjectIDFromHex(sensorId)
	if err2 != nil {
		log.Println("ID non valido:", err2)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error converting sensor ID"})
		return
	}
	sensorList, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": oid})
	if err != nil || len(sensorList) == 0 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching sensor from database"})
		return
	}
	sensor := sensorList[0]
	c.IndentedJSON(http.StatusOK, sensor)
}

func insertSensorConfiguration(c *gin.Context, appConfig *config.AppConfiguration, data bson.M) {
	_, errConfiguration := appConfig.Mongo.InsertData(data)
	if errConfiguration != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting configuration into database"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Sensor inserted successfully"})
}
func editSensorConfiguration(c *gin.Context, appConfig *config.AppConfiguration, sensorId string, data bson.M) {
	oid, err2 := primitive.ObjectIDFromHex(sensorId)
	if err2 != nil {
		log.Println("ID non valido:", err2)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error converting sensor ID"})
		return
	}
	modifiedCount, errUpdate := appConfig.Mongo.UpdateData(
		bson.M{"_id": oid},
		data,
	)
	if errUpdate != nil || modifiedCount == 0 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error updating sensor configuration in database"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Sensor updated successfully"})
}
