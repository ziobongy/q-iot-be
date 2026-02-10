package config

type AppConfiguration struct {
	Mongo  *MongoClient
	Influx *InfluxClient
}

func NewAppConfiguration() *AppConfiguration {
	mongo := NewMongoClient()
	influx := NewInfluxClient()
	return &AppConfiguration{
		Mongo:  mongo,
		Influx: influx,
	}
}
