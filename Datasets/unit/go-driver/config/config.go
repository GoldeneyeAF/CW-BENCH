package main

type Config struct {
	ServerPort	string
	MongoAddr	string
	KafkaBroker	string
	Location	string
}

func NewConfig(port string, mongo string, kafka string, location string) *Config {
	return &Config{ServerPort: port, MongoAddr: mongo, KafkaBroker: kafka, Location: location}
}
