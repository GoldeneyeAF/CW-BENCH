package main

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"math"
)

type LatLngLiteral struct {
	Lat float64 `json:"lat" validate:"required"`
	Lng float64 `json:"lng" validate:"required"`
}

func Distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	radlat1 := math.Pi * lat1 / 180
	radlat2 := math.Pi * lat2 / 180
	theta := lng1 - lng2
	radtheta := math.Pi * theta / 180
	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}
	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515
	dist = dist * 1.609344
	return dist
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


var gaugeOnlineSqlQueries = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "location_service", Name: "online_sql_queries_counter",
})

type LocationRepository struct {
	database *sqlx.DB
}

func NewLocationRepository(m *sqlx.DB) Repository {
	return &LocationRepository{
		database: m,
	}
}

func (lr *LocationRepository) Create(c context.Context, driver *Driver) error {
	sqlStatement := "INSERT INTO drivers (id, latitude, longitude) VALUES ($1, $2, $3)"
	gaugeOnlineSqlQueries.Inc()
	_, err := lr.database.QueryContext(c, sqlStatement, driver.Id, driver.Coordinates.Lat, driver.Coordinates.Lng)
	if err != nil {
		log.Fatal(err)
	}
	gaugeOnlineSqlQueries.Dec()
	return nil
}

func (lr *LocationRepository) FindDrivers(c context.Context, latitude float64, longitude float64, radius float64) ([]Driver, error) {
	gaugeOnlineSqlQueries.Inc()
	rows, err := lr.database.QueryContext(c, "SELECT * FROM drivers")
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)
	if err != nil {
		log.Fatal(err)
	}
	var drivers []Driver
	for rows.Next() {
		var driver Driver
		if err := rows.Scan(&driver.Id, &driver.Coordinates.Lat, &driver.Coordinates.Lng); err != nil {
			return drivers, err
		}
		drivers = append(drivers, driver)
	}
	var nearestDrivers []Driver
	for _, driver := range drivers {
		if Distance(driver.Coordinates.Lat, driver.Coordinates.Lng, latitude, longitude) <= radius {
			nearestDrivers = append(nearestDrivers, driver)
		}
	}
	gaugeOnlineSqlQueries.Dec()
	return nearestDrivers, nil
}

func (lr *LocationRepository) UpdateDriverLocation(c context.Context, id string, latitude float64, longitude float64) error {
	gaugeOnlineSqlQueries.Inc()
	rows, err := lr.database.QueryContext(c, "UPDATE drivers SET latitude = $2, longitude = $3 WHERE id = $1", id, id, longitude)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)
	if err != nil {
		log.Fatal(err)
	}
	gaugeOnlineSqlQueries.Dec()
	return nil
}
