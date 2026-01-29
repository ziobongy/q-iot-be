package service

import (
	"qiot-configuration-service/config"

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
