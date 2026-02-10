package service

import (
	"log"
	"maps"
	"qiot-configuration-service/config"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExperimentService struct {
	AppConfig *config.AppConfiguration
}

func NewExperimentService(appConfig *config.AppConfiguration) *ExperimentService {
	return &ExperimentService{
		AppConfig: appConfig,
	}
}

func (es *ExperimentService) GetAllExperiments() ([]bson.M, error) {
	resultFromMongo, err := es.AppConfig.Mongo.ExecuteSelectionQuery(bson.M{}, "experiments")
	if err != nil {
		return nil, err
	}
	result := make([]bson.M, 0)
	for _, element := range resultFromMongo {
		id := element["_id"]
		delete(element, "devices")
		delete(element, "_id")
		element["id"] = id
		result = append(result, element)
	}
	return result, nil
}
func (es *ExperimentService) GetRawExperimentById(id string) (bson.M, error) {
	oid, err2 := primitive.ObjectIDFromHex(id)
	if err2 != nil {
		return nil, err2
	}
	experimentList, err := es.AppConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": oid}, "experiments")
	if err != nil || len(experimentList) == 0 {
		return nil, err
	}
	experiment := experimentList[0]
	return experiment, nil
}
func (es *ExperimentService) GetCompleteExperimentById(id string) (bson.M, error) {
	result := bson.M{}
	oid, error := primitive.ObjectIDFromHex(id)
	if error != nil {
		log.Println("ID non valido:", error)
		return nil, error
	}
	experimentList, err := es.AppConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": oid}, "experiments")
	if err != nil || len(experimentList) == 0 {
		return nil, err
	}
	experiment := experimentList[0]
	result["devices"] = bson.M{}
	if experiment["devices"] != nil {
		i := 0
		for _, device := range experiment["devices"].(primitive.A) {
			sid, errorSensorId := primitive.ObjectIDFromHex(device.(bson.M)["sensorId"].(string))
			if errorSensorId != nil {
				log.Println("ID non valido:", error)
				return nil, errorSensorId
			}
			sensorList, errSensor := es.AppConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": sid})
			if errSensor != nil || len(sensorList) == 0 {
				return nil, errSensor
			}
			// sensore preso dal db
			sensor := sensorList[0]
			sensorServices := []bson.M{}
			delete(sensor, "_id")
			if device.(primitive.M)["enabledServices"] != nil {
				for _, service := range device.(primitive.M)["enabledServices"].(primitive.A) {
					for _, serviceFromSensor := range sensor["services"].(primitive.A) {
						if serviceFromSensor.(bson.M)["uuid"] == service {
							newCharacteristics := []bson.M{}
							characteristics := serviceFromSensor.(bson.M)["characteristics"].(primitive.A)
							for _, characteristic := range characteristics {
								characteristicMap := characteristic.(bson.M)
								newCharacteristics = append(newCharacteristics, bson.M{
									"name":         characteristicMap["name"],
									"uuid":         characteristicMap["uuid"],
									"structParser": characteristicMap["structParser"],
									"mqttTopic": "qiot/" +
										experiment["_id"].(primitive.ObjectID).Hex() + "/" +
										sensor["name"].(string) + "/" +
										strings.ToLower(strings.Replace(device.(primitive.M)["macAddress"].(string), ":", "", -1)) + "/" +
										strings.ToLower(strings.Replace(characteristicMap["name"].(string), " ", "", -1)),
								})
								log.Println("New characteristics:", newCharacteristics)
							}
							serviceFromSensorMap := serviceFromSensor.(bson.M)
							serviceFromSensorMap["characteristics"] = newCharacteristics
							sensorServices = append(sensorServices, serviceFromSensorMap)
						}
					}
				}
			}
			sensor["services"] = sensorServices
			sensor["address"] = device.(bson.M)["macAddress"]
			if _, ok := device.(bson.M)["dynamicSchema"]; ok {
				sensor["dynamicSchema"] = device.(bson.M)["dynamicSchema"]
				delete(sensor, "dynamicSchema")
				resultJson := sensor["dynamicJson"].(primitive.M)
				keys := maps.Keys(resultJson)
				for key := range keys {
					sensor[key] = resultJson[key]
				}
				delete(sensor, "dynamicJson")
			}
			result["devices"].(bson.M)["sensor_"+strconv.Itoa(i)] = sensor
			i++
		}
	}
	return result, nil
}
func (es *ExperimentService) InsertExperiment(data bson.M) (InsertedID interface{}, err error) {
	inserted, errConfiguration := es.AppConfig.Mongo.InsertData(data, "experiments")
	if errConfiguration != nil {
		return nil, errConfiguration
	}
	completeExperiment, errCompleteExperiment := es.GetCompleteExperimentById(inserted.(primitive.ObjectID).Hex())
	if errCompleteExperiment != nil {
		log.Println("error while retrieving complete experiment:", errCompleteExperiment)
		return nil, errCompleteExperiment
	}
	emqx := NewClientFromEnv()
	errorEmqx := emqx.ProcessYAMLAndSync(completeExperiment["devices"].(bson.M))
	if errorEmqx != nil {
		return nil, errorEmqx
	}
	return inserted, nil
}
func (es *ExperimentService) UpdateExperiment(id string, data bson.M) (int64, error) {
	eid, errorExperimentId := primitive.ObjectIDFromHex(id)
	if errorExperimentId != nil {
		log.Println("ID non valido:", errorExperimentId)
		return 0, errorExperimentId
	}
	inserted, errConfiguration := es.AppConfig.Mongo.UpdateData(bson.M{"_id": eid}, data, "experiments")
	if errConfiguration != nil {
		log.Println("error while inserting:", errConfiguration)
		return 0, errConfiguration
	}
	return inserted, nil
}
