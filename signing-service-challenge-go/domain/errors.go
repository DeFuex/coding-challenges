package domain

import "errors"

var (
	// ErrDeviceNotFound is returned when a device with the given ID doesn't exist
	ErrDeviceNotFound = errors.New("device not found")
	// ErrDeviceAlreadyExists is returned when trying to create a device with an ID that already exists
	ErrDeviceAlreadyExists = errors.New("device already exists")
	// ErrInvalidAlgorithm is returned when an invalid signature algorithm is provided
	ErrInvalidAlgorithm = errors.New("invalid signature algorithm")
)
