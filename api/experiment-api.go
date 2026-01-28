package api

import (
	"log"
	"net/http"
	"qiot-configuration-service/config"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewExperimentAPI(appConfig *config.AppConfiguration, ginEngine *gin.Engine) {
	ginEngine.GET("/experiment", func(c *gin.Context) {
		getExperiments(c, appConfig)
	})
	ginEngine.GET("/experiment/yaml/:id", func(c *gin.Context) {
		getExperimentByIdYaml(c, appConfig, c.Param("id"))
	})
	ginEngine.GET("/experiment/:id", func(c *gin.Context) {
		getRawExperimentById(c, appConfig, c.Param("id"))
	})
	ginEngine.POST("/experiment", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		insertExperiment(c, appConfig, body)
	})
	ginEngine.PUT("/experiment/:experimentId", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updateExperiment(c, appConfig, c.Param("experimentId"), body)
	})
}

func getExperiments(c *gin.Context, appConfig *config.AppConfiguration) {
	resultFromMongo, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{}, "experiments")
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching data from database"})
		return
	}
	result := make([]bson.M, 0)
	for _, element := range resultFromMongo {
		id := element["_id"]
		delete(element, "devices")
		delete(element, "_id")
		element["id"] = id
		result = append(result, element)
	}
	c.IndentedJSON(http.StatusOK, result)
}
func getRawExperimentById(c *gin.Context, appConfig *config.AppConfiguration, experimentId string) {
	oid, err2 := primitive.ObjectIDFromHex(experimentId)
	if err2 != nil {
		log.Println("ID non valido:", err2)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error converting experiment ID"})
		return
	}
	experimentList, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": oid}, "experiments")
	if err != nil || len(experimentList) == 0 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching experiment from database"})
		return
	}
	experiment := experimentList[0]
	c.IndentedJSON(http.StatusOK, experiment)
}
func getExperimentByIdYaml(c *gin.Context, appConfig *config.AppConfiguration, experimentId string) {
	result := bson.M{}
	oid, error := primitive.ObjectIDFromHex(experimentId)
	if error != nil {
		log.Println("ID non valido:", error)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error converting experiment ID"})
		return
	}
	experimentList, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": oid}, "experiments")
	if err != nil || len(experimentList) == 0 {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching experiment from database"})
		return
	}
	experiment := experimentList[0]
	result["devices"] = bson.M{}
	if experiment["devices"] != nil {
		i := 0
		for _, device := range experiment["devices"].(primitive.A) {
			sid, errorSensorId := primitive.ObjectIDFromHex(device.(bson.M)["sensorId"].(string))
			if errorSensorId != nil {
				log.Println("ID non valido:", error)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error converting sensor ID"})
				return
			}
			sensorList, errSensor := appConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": sid})
			if errSensor != nil || len(sensorList) == 0 {
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching sensor from database"})
				return
			}
			// sensore preso dal db
			sensor := sensorList[0]
			sensorServices := []bson.M{}
			delete(sensor, "_id")
			if device.(primitive.M)["enabledServices"] != nil {
				for _, service := range device.(primitive.M)["enabledServices"].(primitive.A) {
					for _, serviceFromSensor := range sensor["services"].(primitive.A) {
						if serviceFromSensor.(bson.M)["uuid"] == service {
							sensorServices = append(sensorServices, serviceFromSensor.(bson.M))
						}
					}
				}
			}
			sensor["services"] = sensorServices
			result["devices"].(bson.M)["sensor_"+strconv.Itoa(i)] = sensor
			i++
		}
	}
	c.YAML(http.StatusOK, gin.H(result))
}

func insertExperiment(c *gin.Context, appConfig *config.AppConfiguration, data bson.M) {
	_, errConfiguration := appConfig.Mongo.InsertData(data, "experiments")
	if errConfiguration != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting configuration into database"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Experiment inserted successfully"})
}
func updateExperiment(c *gin.Context, appConfig *config.AppConfiguration, experimentId string, data bson.M) {
	eid, errorExperimentId := primitive.ObjectIDFromHex(experimentId)
	if errorExperimentId != nil {
		log.Println("ID non valido:", errorExperimentId)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error converting sensor ID"})
		return
	}
	_, errConfiguration := appConfig.Mongo.UpdateData(bson.M{"_id": eid}, data, "experiments")
	if errConfiguration != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting configuration into database"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Experiment inserted successfully"})
}
