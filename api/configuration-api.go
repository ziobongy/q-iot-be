package api

import (
	"qiot-configuration-service/config"

	"github.com/gin-gonic/gin"
)

type ConfigurationAPI struct {
	app        *config.AppConfiguration
	ginContext *gin.Context
}

func NewConfigurationAPI(appConfig *config.AppConfiguration, ginEngine *gin.Engine) {
	// ginEngine.GET("/config", func(c *gin.Context) {
	// 	getData(c, appConfig)
	// })
	// ginEngine.GET("/config/experiments", func(c *gin.Context) {
	// 	getExperiments(c, appConfig)
	// })
	// ginEngine.GET("/config/experiments/:experimentId", func(c *gin.Context) {
	// 	getExperiment(c, appConfig, c.Param("experimentId"))
	// })
	// ginEngine.POST("/config", func(c *gin.Context) {
	// 	var body bson.M
	// 	if err := c.ShouldBindJSON(&body); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// 	insertData(c, appConfig, body)
	// })
}

// func getData(c *gin.Context, appConfig *config.AppConfiguration) {
// 	result, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{})
// 	if err != nil {
// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching data from database"})
// 		return
// 	}
// 	c.IndentedJSON(http.StatusOK, result)
// }
//
// func getExperiments(c *gin.Context, appConfig *config.AppConfiguration) {
// 	result, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{}, "experiments")
// 	if err != nil {
// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching data from database"})
// 		return
// 	}
// 	toReturn := make([]bson.M, 0)
// 	for _, element := range result {
// 		id := element["_id"]
// 		toReturn = append(toReturn, bson.M{
// 			"id":          id,
// 			"description": element["description"],
// 		})
// 	}
// 	c.IndentedJSON(
// 		http.StatusOK,
// 		toReturn,
// 	)
// }
//
// func getExperiment(c *gin.Context, appConfig *config.AppConfiguration, experimentId string) {
// 	oid, error := primitive.ObjectIDFromHex(experimentId)
// 	if error != nil {
// 		log.Println("ID non valido:", error)
// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error converting experiment ID"})
// 		return
// 	}
// 	result, err := appConfig.Mongo.ExecuteSelectionQuery(bson.M{"experiment": oid}, "configurations")
// 	if err != nil {
// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching data from database"})
// 		return
// 	}
// 	sensorParsed := bson.M{}
// 	for i := range result {
// 		delete(result[i], "experiment")
// 		delete(result[i], "_id")
// 		sensorParsed["sensor_"+strconv.Itoa(i)] = result[i]
// 	}
//
// 	c.YAML(http.StatusOK, gin.H{"devices": sensorParsed})
// }
//
// func insertData(c *gin.Context, appConfig *config.AppConfiguration, data bson.M) {
// 	var experimentDescription string
// 	var resultExperiment interface{}
// 	var errExperiment error
// 	if data != nil && data["description"] != nil {
// 		experimentDescription = data["description"].(string)
// 		resultExperiment, errExperiment = appConfig.Mongo.InsertData(
// 			bson.M{
// 				"description": experimentDescription,
// 			},
// 			"experiments",
// 		)
// 		if errExperiment != nil {
// 			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting experiment into database"})
// 			return
// 		}
// 	} else {
// 		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "No experiment description provided"})
// 		return
// 	}
// 	for _, sensor := range data["sensors"].([]interface{}) {
// 		sensor.(map[string]interface{})["experiment"] = resultExperiment
// 		_, errConfiguration := appConfig.Mongo.InsertData(
// 			sensor.(map[string]interface{}),
// 		)
// 		if errConfiguration != nil {
// 			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting configuration into database"})
// 			return
// 		}
// 	}
// 	c.IndentedJSON(http.StatusOK, gin.H{"id": resultExperiment})
// }
//
