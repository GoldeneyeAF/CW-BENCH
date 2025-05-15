package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type LatLngLiteral struct {
	Lat float64 `json:"lat" validate:"required"`
	Lng float64 `json:"lng" validate:"required"`
}

type Driver struct {
	Id          string
	Coordinates LatLngLiteral
}

type Repository interface {
	Create(c context.Context, driver *Driver) error
	FindDrivers(c context.Context, latitude float64, longitude float64, radius float64) ([]Driver, error)
	UpdateDriverLocation(c context.Context, id string, latitude float64, longitude float64) error
}

type Usecase interface {
	Create(driver *Driver) error
	FindDrivers(latitude float64, longitude float64, radius float64) ([]Driver, error)
	UpdateDriverLocation(id string, latitude float64, longitude float64) error
}

type LocationController struct {
	LocationUsecase Usecase
}

var (
	counterGetDrivers = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "location_service", Name: "execution_of_get_drivers_handler",
	})
	counterPostDriverLocation = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "location_service", Name: "execution_of_post_driver_location_handler",
	})
)

func (dc *LocationController) GetDrivers(w http.ResponseWriter, r *http.Request) {
	counterGetDrivers.Inc()
	if r.Method != http.MethodGet { // проверяем GET ли метод
		http.Error(w, "Use of wrong HTTP-method", http.StatusMethodNotAllowed)
		return
	}
	lat := r.URL.Query().Get("lat") // проверяем каждый query-параметр на наличие
	if lat == "" {
		http.Error(w, "Empty query parameter lat", http.StatusBadRequest)
		return
	}
	lng := r.URL.Query().Get("lng")
	if lng == "" {
		http.Error(w, "Empty query parameter lng", http.StatusBadRequest)
		return
	}
	radius := r.URL.Query().Get("radius")
	if radius == "" {
		http.Error(w, "Empty query parameter radius", http.StatusBadRequest)
		return
	}
	latFloat, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		http.Error(w, "Error occurred converting query param to float64", http.StatusInternalServerError)
		return
	}
	lngFloat, err := strconv.ParseFloat(lng, 64)
	if err != nil {
		http.Error(w, "Error occurred converting query param to float64", http.StatusInternalServerError)
		return
	}
	radiusFloat, err := strconv.ParseFloat(radius, 64)
	if err != nil {
		http.Error(w, "Error occurred converting query param to float64", http.StatusInternalServerError)
		return
	}
	foundDrivers, err := []float64{latFloat, lngFloat, radiusFloat}, nil
	if err != nil {
		http.Error(w, "Error occurred during finding drivers", http.StatusInternalServerError)
		return
	}
	if len(foundDrivers) == 0 {
		http.Error(w, "Drivers not found", http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(foundDrivers)
	if err != nil {
		http.Error(w, "Failed to make output array of drivers", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (dc *LocationController) PostDriverLocation(w http.ResponseWriter, r *http.Request) {
	counterPostDriverLocation.Inc()
	if r.Method != http.MethodPost { // проверяем POST ли метод
		http.Error(w, "Use of wrong HTTP-method", http.StatusMethodNotAllowed)
		return
	}
	vars := mux.Vars(r)
	_, check := vars["driver_id"]
	if !check {
		http.Error(w, "Id is missing in parameters", http.StatusBadRequest)
		return
	}
	var coordinatesToUpdate LatLngLiteral
	err := json.NewDecoder(r.Body).Decode(&coordinatesToUpdate)
	if err != nil { // проверяем получилось ли открыть json
		http.Error(w, "Failed to decode json", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(coordinatesToUpdate)
	if err != nil { // проверяем, что поля json действительно соответствуют полям структуры из домена
		http.Error(w, "Failed to validate struct", http.StatusBadRequest)
		return
	}
	err = nil
	if err != nil {
		http.Error(w, "Failed to update location", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/text")
	_, err = w.Write([]byte("Success operation\n"))
	if err != nil {
		log.Fatal(err)
	}
}
