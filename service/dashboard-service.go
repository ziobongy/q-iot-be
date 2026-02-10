package service

import (
	"qiot-configuration-service/config"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DashboardService struct {
	AppConfig         *config.AppConfiguration
	ExperimentService *ExperimentService
}
type ElementToQuery struct {
	Bucket        string
	SensorName    string
	DeviceAddress string
	Measurement   string
	Field         string
}

func NewDashboardService(appConfig *config.AppConfiguration) *DashboardService {
	return &DashboardService{
		AppConfig:         appConfig,
		ExperimentService: NewExperimentService(appConfig),
	}
}
func (ds *DashboardService) GetDashboardData(experimentId string, characteristicId string) ([]bson.M, error) {
	//devo fare una query a influx per deviceAddress = deviceAddress e _measurement = measurement
	experiment, err := ds.ExperimentService.GetCompleteExperimentById(experimentId)
	if err != nil {
		return nil, err
	}
	elementToQuery := []ElementToQuery{}
	devices := experiment["devices"].(primitive.M)
	for _, device := range devices {
		deviceMap := device.(primitive.M)
		name := strings.ToLower(deviceMap["name"].(string))
		for _, service := range deviceMap["services"].([]primitive.M) {
			if service["uuid"].(string) != characteristicId {
				continue
			}
			for _, characteristic := range service["characteristics"].(primitive.A) {
				characteristicMap := characteristic.(primitive.M)
				characteristicName := characteristicMap["name"].(string)
				finalName := name + "_" + strings.ToLower(strings.Replace(characteristicName, " ", "", -1))
				structParser := characteristicMap["structParser"].(primitive.M)
				for _, field := range structParser["fields"].(primitive.A) {
					fieldMap := field.(primitive.M)
					element := ElementToQuery{
						Bucket:        "iotproject_bucket",
						SensorName:    finalName,
						Measurement:   finalName,
						DeviceAddress: deviceMap["address"].(string),
						Field:         fieldMap["name"].(string),
					}
					elementToQuery = append(elementToQuery, element)
				}
			}
		}
	}
	result := []bson.M{}
	for _, element := range elementToQuery {
		categories, data, err := ds.AppConfig.Influx.ExecuteQuery(element.Bucket, element.DeviceAddress, element.Measurement, element.Field)
		if err != nil {
			return nil, err
		}
		result = append(
			result,
			bson.M{
				"id":         element.SensorName + element.Measurement + element.Field,
				"sensorName": element.SensorName + " - " + element.Measurement + " - " + element.Field,
				"categories": categories,
				"data":       data,
			},
		)
	}
	return result, nil
}
