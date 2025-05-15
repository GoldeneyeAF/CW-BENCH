package main

import (
	"context"
	"time"
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

type LocationUsecase struct {
	locationRepository Repository
	contextTimeout     time.Duration
}

func NewLocationUsecase(driverRepository Repository, timeout time.Duration) Usecase {
	return &LocationUsecase{
		locationRepository: driverRepository,
		contextTimeout:     timeout,
	}
}

func (du *LocationUsecase) Create(driver *Driver) error {
	ctx, cancel := context.WithTimeout(context.Background(), du.contextTimeout)
	defer cancel()
	return du.locationRepository.Create(ctx, driver)
}

func (du *LocationUsecase) FindDrivers(latitude float64, longitude float64, radius float64) ([]Driver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), du.contextTimeout)
	defer cancel()
	return du.locationRepository.FindDrivers(ctx, latitude, longitude, radius)
}

func (du *LocationUsecase) UpdateDriverLocation(id string, latitude float64, longitude float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), du.contextTimeout)
	defer cancel()
	return du.locationRepository.UpdateDriverLocation(ctx, id, latitude, longitude)
}
