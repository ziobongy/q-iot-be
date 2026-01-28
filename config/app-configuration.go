package config

type AppConfiguration struct {
	Mongo *MongoClient
}

func NewAppConfiguration() *AppConfiguration {
	mongo := NewMongoClient()
	return &AppConfiguration{
		Mongo: mongo,
	}
}
