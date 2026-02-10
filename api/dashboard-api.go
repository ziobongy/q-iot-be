package api

import (
	"net/http"
	"qiot-configuration-service/config"
	"qiot-configuration-service/service"

	"github.com/gin-gonic/gin"
)

func NewDashboardAPI(appConfig *config.AppConfiguration, ginEngine *gin.Engine) {
	dashboardService := service.NewDashboardService(appConfig)
	ginEngine.GET("/dashboard/:experimentId/device/:sensorId", func(c *gin.Context) {
		dashboardForSensor(c, dashboardService, c.Param("experimentId"), c.Param("sensorId"))
	})
}

func dashboardForSensor(c *gin.Context, es *service.DashboardService, experimentId string, characteristicId string) {
	result, _ := es.GetDashboardData(experimentId, characteristicId)
	c.IndentedJSON(http.StatusOK, result)

}
