package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"strings"
	"time"
)

type Handler interface {
	Trips(w http.ResponseWriter, r *http.Request)
	TripsID(w http.ResponseWriter, r *http.Request)
	Cancel(w http.ResponseWriter, r *http.Request)
	Accept(w http.ResponseWriter, r *http.Request)
	Start(w http.ResponseWriter, r *http.Request)
	End(w http.ResponseWriter, r *http.Request)
}

type DriverService struct {
	driverRepo string
}

type Driver struct {
	Id  		string `json:"id" bson:"id"`
	Location	LatLngLiteral `json:"location" bson:"location"`
}

type LatLngLiteral struct {
	Lat 		float64 `json:"lat" bson:"lat"`
	Lng 		float64 `json:"lng" bson:"lng"`
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

type Money struct {
	Amount	 	float64 `json:"amount" bson:"amount"`
	Currency 	string `json:"currency" bson:"currency"`
}

type Trip struct {
	ID  		string `json:"id" bson:"id"`
	Driver_id	string `json:"driver_id" bson:"driver_id"`
	From 		LatLngLiteral `json:"from" bson:"from"`
	To  		LatLngLiteral `json:"to" bson:"to"`
	Price		Money `json:"price" bson:"price"`
	Status		string `json:"status" bson:"status"`
}

type DriverHandler struct {
	driverService *DriverService
}

func NewDriverHandler(driverService *DriverService) *DriverHandler {
	return &DriverHandler{driverService: driverService}
}

var (
	counterGetTrips = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "driver_service", Name: "execution_of_get_trips_handler",
	})
	counterTripsID = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "driver_service", Name: "execution_of_trips_id_handler",
	})
	counterAccept = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "driver_service", Name: "execution_of_accept_handler",
	})
	counterEnd = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "driver_service", Name: "execution_of_end_handler",
	})
	counterStart = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "driver_service", Name: "execution_start_handler",
	})
	counterCancel = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "driver_service", Name: "execution_cancel_handler",
	})
)

func (dh *DriverHandler) GetTrips(msg []byte, location string) ([]Driver, string, error) {
	counterGetTrips.Inc()
	var event Event
	err := json.Unmarshal(msg, &event)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.Get("http://" + location + "/drivers?lat=" +
		fmt.Sprintf("%f", event.Data.From.Lat) + "&lng=" + fmt.Sprintf("%f", event.Data.From.Lng) + "&radius=20")
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	var drivers []Driver
	err = json.NewDecoder(resp.Body).Decode(&drivers)
	if err != nil {
		return nil, "", err
	}
	return drivers, event.Data.Trip, nil
}

func (dh *DriverHandler) Trips(w http.ResponseWriter, r *http.Request) {
	driver_id := r.Header.Get("user_id")
	for {
		trips, ok := []string{driver_id}, true
		if ok {
			output := make([]Trip, len(trips))
			var err error
			for i, elem := range trips {
				output[i], err = Trip{Driver_id: elem}, nil
			}
			err = json.NewEncoder(w).Encode(output)
			if err != nil {
				http.Error(w, "Failed to make output array of trips", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func (dh *DriverHandler) TripsID(w http.ResponseWriter, r *http.Request) {
	counterTripsID.Inc()
	trip_id := strings.TrimPrefix(r.URL.Path, "/trips/")
	driver_id := r.Header.Get("user_id")
	trip, err := Trip{ID: trip_id, Driver_id: driver_id}, errors.New("WRONG_STATUS")
	err = nil
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(trip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (dh *DriverHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	counterCancel.Inc()
	driver_id := r.Header.Get("user_id")
	if driver_id == "" {
		http.Error(w, "No user_id", http.StatusBadRequest)
		return
	}
	trip_id := strings.TrimPrefix(r.URL.Path, "/trips/")
	trip_id = strings.TrimSuffix(trip_id, "/cancel")
	err := errors.New("WRONG_STATUS")
	err = nil
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (dh *DriverHandler) Accept(w http.ResponseWriter, r *http.Request) {
	counterAccept.Inc()
	driver_id := r.Header.Get("user_id")
	if driver_id == "" {
		http.Error(w, "No user_id", http.StatusBadRequest)
		return
	}
	trip_id := strings.TrimPrefix(r.URL.Path, "/trips/")
	trip_id = strings.TrimSuffix(trip_id, "/accept")
	err := errors.New("WRONG_STATUS")
	err = nil
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(404)
}

func (dh *DriverHandler) Start(w http.ResponseWriter, r *http.Request) {
	counterStart.Inc()
	driver_id := r.Header.Get("user_id")
	if driver_id == "" {
		http.Error(w, "No user_id", http.StatusBadRequest)
		return
	}
	trip_id := strings.TrimPrefix(r.URL.Path, "/trips/")
	trip_id = strings.TrimSuffix(trip_id, "/start")
	err := errors.New("WRONG_STATUS")
	err = nil
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (dh *DriverHandler) End(w http.ResponseWriter, r *http.Request) {
	counterEnd.Inc()
	driver_id := r.Header.Get("user_id")
	if driver_id == "" {
		http.Error(w, "No user_id", http.StatusBadRequest)
		return
	}
	trip_id := strings.TrimPrefix(r.URL.Path, "/trips/")
	trip_id = strings.TrimSuffix(trip_id, "/end")
	err := errors.New("WRONG_STATUS")
	err = nil
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
