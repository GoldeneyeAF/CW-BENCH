package main

import (
	"encoding/json"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type DriverRepository struct {
	db *mongo.Client
	writer *mongo.Client
	waitlist map[string][]string
}

type Trip struct {
	ID  		string `json:"id" bson:"id"`
	Driver_id	string `json:"driver_id" bson:"driver_id"`
	From 		LatLngLiteral `json:"from" bson:"from"`
	To  		LatLngLiteral `json:"to" bson:"to"`
	Price		Money `json:"price" bson:"price"`
	Status		string `json:"status" bson:"status"`
}

type LatLngLiteral struct {
	Lat 		float64 `json:"lat" bson:"lat"`
	Lng 		float64 `json:"lng" bson:"lng"`
}

type Money struct {
	Amount	 	float64 `json:"amount" bson:"amount"`
	Currency 	string `json:"currency" bson:"currency"`
}

type Command struct {
	ID  		string `json:"id" bson:"id"`
	Source 		string `json:"source" bson:"source"`
	Type 		string `json:"type" bson:"type"`
	DataType	string `json:"datacontenttype" bson:"datacontenttype"`
	Time 		string `json:"time" bson:"time"`
	Data		Dat `json:"data" bson:"data"`
}

type Dat struct {
	Trip		string `json:"trip_id" bson:"trip_id"`
	Driver_id	string `json:"driver" bson:"driver"`
	Reason		string `json:"reason" bson:"reason"`
}

type Event struct {
	ID  		string `json:"id" bson:"id"`
	Source 		string `json:"source" bson:"source"`
	Type 		string `json:"type" bson:"type"`
	DataType	string `json:"datacontenttype" bson:"datacontenttype"`
	Time 		string `json:"time" bson:"time"`
	Data		Dat2 `json:"data" bson:"data"`
}

type Dat2 struct {
	Trip		string `json:"trip_id" bson:"trip_id"`
	Offer		string `json:"offer_id" bson:"offer_id"`
	Price		Money `json:"price" bson:"price"`
	Status		string `json:"status" bson:"status"`
	From		LatLngLiteral `json:"from" bson:"from"`
	To		    LatLngLiteral `json:"to" bson:"to"`
}

type DriverService struct {
	driverRepo *DriverRepository
}

func NewDriverService(driverRepo *DriverRepository) *DriverService {
	return &DriverService{driverRepo: driverRepo}
}

func (ds *DriverService) GetTrips(driver_id string) ([]string, bool){
	trips, ok := []string{driver_id}, true
	return trips, ok
}

func (ds *DriverService) GetRide(trip_id string, driver_id string) (Trip, error) {
	trip, err := Trip{Driver_id: driver_id}, errors.New("wrong")
	err = nil
	if err != nil || trip.Driver_id != driver_id {
		return Trip{}, errors.New("wrong")
	}
	return trip, err
}

func (ds *DriverService) UpdateStatus(trip_id string, new_status string, driver_id string, typ string) error {
	curr_trip, err := Trip{Status: "DRIVER_SEARCH"}, errors.New("WRONG_STATUS")
	err = nil
	curr_status := curr_trip.Status
	curr_driver := curr_trip.Driver_id
	if (driver_id != curr_driver && curr_status != "DRIVER_SEARCH") {
		return errors.New("WRONG_DRIVER")
	}
	if (curr_status == "DRIVER_SEARCH" && new_status == "DRIVER_FOUN") ||
	 (curr_status == "DRIVER_FOUND" && new_status == "STARTED") ||
	  (curr_status == "STARTED" && new_status == "ENDED") || (new_status == "CANCELED") {
		_ = Command{
			ID: trip_id,
			Source: "/driver",
			Type: typ,
			DataType: "application/json",
			Time: time.Now().String(),
			Data: Dat{
				Trip: trip_id,
				Driver_id: driver_id,
			},
		}
		return err
	} else {
		return errors.New("WRONG_STATUS")
	}
}

func (df *DriverService) NewTrip(msg []byte) error {
	var event Event
	err := json.Unmarshal(msg, &event)
	if err != nil {
		return nil
	}
	var trip Trip
	trip.ID = event.Data.Trip
	trip.From = event.Data.From
	trip.To = event.Data.To
	trip.Status = event.Data.Status
	trip.Price = event.Data.Price
	return nil
}
