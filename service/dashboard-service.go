package service

import (
	"qiot-configuration-service/config"
	"regexp"
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
	nonAlpha := regexp.MustCompile(`[^a-z0-9]`)
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
		deviceShort := strings.ToLower(getString(deviceMap, "shortName"))
		for _, service := range deviceMap["services"].([]primitive.M) {
			if service["uuid"].(string) != characteristicId {
				continue
			}
			for _, characteristic := range service["characteristics"].([]primitive.M) {
				characteristicMap := characteristic
				characteristicName := strings.ToLower(strings.Replace(characteristicMap["name"].(string), " ", "", -1))
				clean := nonAlpha.ReplaceAllString(characteristicName, "")
				if deviceShort == "" || clean == "" {
					continue
				}
				measureName := deviceShort + "_" + clean
				finalName := name + "_" + characteristicName
				structParser := characteristicMap["structParser"].(primitive.M)
				for _, field := range structParser["fields"].(primitive.A) {
					fieldMap := field.(primitive.M)
					element := ElementToQuery{
						Bucket:        "iotproject_bucket",
						SensorName:    finalName,
						Measurement:   measureName,
						DeviceAddress: deviceMap["address"].(string),
						Field:         fieldMap["name"].(string),
					}
					elementToQuery = append(elementToQuery, element)
				}
			}
		}

		if wbs, ok := deviceMap["movesense_whiteboard"]; ok {
			if measures, ok := wbs.(primitive.M)["measures"]; ok {
				for _, measure := range measures.([]primitive.M) {
					mname := strings.ToLower(getString(measure, "name"))
					clean := nonAlpha.ReplaceAllString(mname, "")
					if deviceShort == "" || clean == "" {
						continue
					}
					measureName := deviceShort + "_" + clean
					if jp, ok := measure["jsonPayloadParser"].(bson.M); ok {
						fields := []bson.M{}
						if farr, ok := jp["fields"].(bson.A); ok {
							for _, fi := range farr {
								if fm, ok := fi.(bson.M); ok {
									fields = append(fields, fm)
								}
							}
						}
						element := ElementToQuery{
							Bucket:        "iotproject_bucket",
							SensorName:    mname,
							Measurement:   strings.ToLower(strings.Replace(measureName, " ", "", -1)),
							DeviceAddress: deviceMap["address"].(string),
							Field:         "",
						}
						elementToQuery = append(elementToQuery, element)
					} else if ja, ok := measure["jsonArrayParser"].(bson.M); ok {
						fields := []bson.M{}
						if farr, ok := ja["fields"].(bson.A); ok {
							for _, fi := range farr {
								if fm, ok := fi.(bson.M); ok {
									fields = append(fields, fm)
								}
							}
						}
						for _, f := range fields {
							fname := getString(f, "name")
							element := ElementToQuery{
								Bucket:        "iotproject_bucket",
								SensorName:    mname,
								Measurement:   strings.ToLower(strings.Replace(measureName, " ", "", -1)),
								DeviceAddress: deviceMap["address"].(string),
								Field:         fname,
							}
							elementToQuery = append(elementToQuery, element)
						}
					} else if smp, ok := measure["SingleMeasurementParser"].(bson.A); ok {
						for _, fi := range smp {
							if fm, ok := fi.(bson.M); ok {
								fname := getString(fm, "name")
								element := ElementToQuery{
									Bucket:        "iotproject_bucket",
									SensorName:    mname,
									Measurement:   strings.ToLower(strings.Replace(measureName, " ", "", -1)),
									DeviceAddress: deviceMap["address"].(string),
									Field:         fname,
								}
								elementToQuery = append(elementToQuery, element)
							}
						}
					}
				}
			}
		}
	}
	result := []bson.M{}
	for _, element := range elementToQuery {
		categories, data, err := ds.AppConfig.Influx.ExecuteQuery(experimentId, element.Bucket, element.DeviceAddress, element.Measurement, element.Field)
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
