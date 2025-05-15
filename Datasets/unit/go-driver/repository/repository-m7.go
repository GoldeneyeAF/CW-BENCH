package main

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Trip struct {
	ID        string        `json:"id" bson:"id"`
	Driver_id string        `json:"driver_id" bson:"driver_id"`
	From      LatLngLiteral `json:"from" bson:"from"`
	To        LatLngLiteral `json:"to" bson:"to"`
	Price     Money         `json:"price" bson:"price"`
	Status    string        `json:"status" bson:"status"`
}

type LatLngLiteral struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

type Money struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}

type Command struct {
	ID       string `json:"id" bson:"id"`
	Source   string `json:"source" bson:"source"`
	Type     string `json:"type" bson:"type"`
	DataType string `json:"datacontenttype" bson:"datacontenttype"`
	Time     string `json:"time" bson:"time"`
	Data     Dat    `json:"data" bson:"data"`
}

type Dat struct {
	Trip      string `json:"trip_id" bson:"trip_id"`
	Driver_id string `json:"driver" bson:"driver"`
	Reason    string `json:"reason" bson:"reason"`
}

type DriverRepository struct {
	db       *mongo.Client
	writer   *mongo.Client
	waitlist map[string][]string
}

func NewDriverRepository(db *mongo.Client, writer *mongo.Client) *DriverRepository {
	return &DriverRepository{db: db, writer: writer, waitlist: make(map[string][]string)}
}

func (r *DriverRepository) InsertTrip(driver_id string, new_trip string) {
	r.waitlist[driver_id] = append(r.waitlist[driver_id], new_trip)
}

func (r *DriverRepository) GetTrips(driver_id string) ([]string, bool) {
	trips, ok := r.waitlist[driver_id]
	delete(r.waitlist, driver_id)
	return trips, ok
}

func (r *DriverRepository) Create(trip Trip) error {
	col := r.db.Database("mainframe").Collection("trips")
	_, err := col.InsertOne(context.TODO(), trip)
	return err
}

func (r *DriverRepository) Update(trip_id string, status string) error {
	col := r.db.Database("mainframe").Collection("trips")
	filter := bson.M{
		"id": trip_id,
	}
	update := bson.M{
		"$set": bson.M{"status": status},
	}
	_, err := col.UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *DriverRepository) Find(trip_id string) (Trip, error) {
	col := r.db.Database("mainframe").Collection("trips")
	filter := bson.M{
		"id": trip_id,
	}
	var result Trip
	err := col.FindOne(context.TODO(), filter).Decode(&result)
	return result, err
}

func (r *DriverRepository) SendCommand(data Command) error {
	_, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return err
}
