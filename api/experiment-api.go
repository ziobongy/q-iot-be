package api

import (
	"net/http"
	"qiot-configuration-service/config"
	"qiot-configuration-service/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func NewExperimentAPI(appConfig *config.AppConfiguration, ginEngine *gin.Engine) {
	es := service.NewExperimentService(appConfig)
	ginEngine.GET("/experiment", func(c *gin.Context) {
		getExperiments(c, es)
	})
	ginEngine.GET("/experiment/yaml/:id", func(c *gin.Context) {
		getExperimentByIdYaml(c, es, c.Param("id"))
	})
	ginEngine.GET("/experiment/json/:id", func(c *gin.Context) {
		getExperimentByIdJson(c, es, c.Param("id"))
	})
	ginEngine.GET("/experiment/:id", func(c *gin.Context) {
		getRawExperimentById(c, es, c.Param("id"))
	})
	ginEngine.POST("/experiment", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		insertExperiment(c, es, body)
	})
	ginEngine.PUT("/experiment/:experimentId", func(c *gin.Context) {
		var body bson.M
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updateExperiment(c, es, c.Param("experimentId"), body)
	})
}

func getExperiments(c *gin.Context, es *service.ExperimentService) {
	result, err := es.GetAllExperiments()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching data from database"})
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}
func getRawExperimentById(c *gin.Context, es *service.ExperimentService, experimentId string) {
	result, err := es.GetRawExperimentById(experimentId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error fetching experiment from database"})
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}
func getExperimentByIdYaml(c *gin.Context, es *service.ExperimentService, experimentId string) {
	result, err := es.GetCompleteExperimentById(experimentId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error while fetching experiment from database"})
		return
	}
	c.YAML(http.StatusOK, gin.H(result))
}
func getExperimentByIdJson(c *gin.Context, es *service.ExperimentService, experimentId string) {
	result, err := es.GetCompleteExperimentById(experimentId)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error while fetching experiment from database"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H(result))
}

func insertExperiment(c *gin.Context, es *service.ExperimentService, data bson.M) {
	_, err := es.InsertExperiment(data)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error while inserting experiment"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Experiment inserted successfully"})
}
func updateExperiment(c *gin.Context, es *service.ExperimentService, experimentId string, data bson.M) {
	_, err := es.UpdateExperiment(experimentId, data)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error while updating experiment"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"result": "Experiment updated successfully"})
}
