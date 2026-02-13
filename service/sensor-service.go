package service

import (
	"log"
	"maps"
	"qiot-configuration-service/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SensorService struct {
	AppConfig *config.AppConfiguration
}

func NewSensorService(appConfig *config.AppConfiguration) *SensorService {
	return &SensorService{
		AppConfig: appConfig,
	}
}

func (ss *SensorService) GetAllSensors() ([]bson.M, error) {
	result, err := ss.AppConfig.Mongo.ExecuteSelectionQuery(bson.M{})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ss *SensorService) GetSensorById(id string) (bson.M, error) {
	oid, err2 := primitive.ObjectIDFromHex(id)
	if err2 != nil {
		log.Println("ID non valido:", err2)
		return nil, err2
	}
	sensorList, err := ss.AppConfig.Mongo.ExecuteSelectionQuery(bson.M{"_id": oid})
	if err != nil || len(sensorList) == 0 {
		return nil, err
	}
	sensor := sensorList[0]
	if _, ok := sensor["dynamicSchema"]; ok {
		//delete(sensor, "dynamicSchema")
		resultJson := sensor["dynamicJson"].(primitive.M)
		keys := maps.Keys(resultJson)
		for key := range keys {
			sensor[key] = resultJson[key]
		}
		//delete(sensor, "dynamicJson")
	}
	return sensor, nil
}
func (ss *SensorService) InsertSensor(data bson.M) (InsertedId interface{}, err error) {
	inserted, errConfiguration := ss.AppConfig.Mongo.InsertData(data)
	if errConfiguration != nil {
		return nil, errConfiguration
	}
	return inserted, nil
}
func (ss *SensorService) EditSensorConfiguration(sensorId string, data bson.M) (int64, error) {
	oid, err2 := primitive.ObjectIDFromHex(sensorId)
	if err2 != nil {
		log.Println("ID non valido:", err2)
		return -1, err2
	}
	modifiedCount, errUpdate := ss.AppConfig.Mongo.UpdateData(
		bson.M{"_id": oid},
		data,
	)
	if errUpdate != nil || modifiedCount == 0 {
		return -1, errUpdate
	}
	return modifiedCount, nil
}
