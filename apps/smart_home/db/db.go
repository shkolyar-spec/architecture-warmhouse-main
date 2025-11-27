package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"smarthome/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents the database connection
type DB struct {
	Pool *pgxpool.Pool
}

// GetSensors returns all sensors
func (db *DB) GetSensors(ctx context.Context) ([]models.Sensor, error) {
	query := `
		SELECT id, name, type, location, value, unit, status, last_updated, created_at
		FROM sensors
		ORDER BY id
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors: %w", err)
	}
	defer rows.Close()

	var sensors []models.Sensor
	for rows.Next() {
		var s models.Sensor
		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Type,
			&s.Location,
			&s.Value,
			&s.Unit,
			&s.Status,
			&s.LastUpdated,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning sensor row: %w", err)
		}
		sensors = append(sensors, s)
	}

	return sensors, nil
}

// GetSensorByID returns a sensor by its ID
func (db *DB) GetSensorByID(ctx context.Context, id int) (models.Sensor, error) {
	query := `
		SELECT id, name, type, location, value, unit, status, last_updated, created_at
		FROM sensors
		WHERE id = $1
	`

	var s models.Sensor
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.Name,
		&s.Type,
		&s.Location,
		&s.Value,
		&s.Unit,
		&s.Status,
		&s.LastUpdated,
		&s.CreatedAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("sensor not found")
	}

	return s, nil
}

// CreateSensor inserts a new sensor into the database
func (db *DB) CreateSensor(ctx context.Context, sc models.SensorCreate) (models.Sensor, error) {
	query := `
		INSERT INTO sensors (name, type, location, value, unit, status, last_updated, created_at)
		VALUES ($1, $2, $3, $4, $5, 'inactive', $6, $7)
		RETURNING id, name, type, location, value, unit, status, last_updated, created_at
	`

	now := time.Now().UTC()
	var s models.Sensor
	err := db.Pool.QueryRow(ctx, query,
		sc.Name,
		sc.Type,
		sc.Location,
		0.0,
		sc.Unit,
		now,
		now,
	).Scan(
		&s.ID,
		&s.Name,
		&s.Type,
		&s.Location,
		&s.Value,
		&s.Unit,
		&s.Status,
		&s.LastUpdated,
		&s.CreatedAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error creating sensor: %w", err)
	}

	return s, nil
}

// UpdateSensor updates a sensor's fields
func (db *DB) UpdateSensor(ctx context.Context, id int, su models.SensorUpdate) (models.Sensor, error) {
	current, err := db.GetSensorByID(ctx, id)
	if err != nil {
		return models.Sensor{}, err
	}

	if su.Name != "" {
		current.Name = su.Name
	}
	if su.Type != "" {
		current.Type = su.Type
	}
	if su.Location != "" {
		current.Location = su.Location
	}
	if su.Unit != "" {
		current.Unit = su.Unit
	}
	if su.Status != "" {
		current.Status = su.Status
	}
	if su.Value != nil {
		current.Value = *su.Value
	}

	query := `
		UPDATE sensors
		SET name = $1, type = $2, location = $3, value = $4, unit = $5, status = $6, last_updated = $7
		WHERE id = $8
		RETURNING id, name, type, location, value, unit, status, last_updated, created_at
	`

	var updated models.Sensor
	err = db.Pool.QueryRow(ctx, query,
		current.Name,
		current.Type,
		current.Location,
		current.Value,
		current.Unit,
		current.Status,
		time.Now().UTC(),
		id,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Type,
		&updated.Location,
		&updated.Value,
		&updated.Unit,
		&updated.Status,
		&updated.LastUpdated,
		&updated.CreatedAt,
	)
	if err != nil {
		return models.Sensor{}, fmt.Errorf("error updating sensor: %w", err)
	}

	return updated, nil
}

// UpdateSensorValue updates only the value and status of a sensor
func (db *DB) UpdateSensorValue(ctx context.Context, id int, value float64, status string) error {
	query := `
		UPDATE sensors
		SET value = $1, status = $2, last_updated = $3
		WHERE id = $4
	`

	result, err := db.Pool.Exec(ctx, query, value, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("error updating sensor value: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("sensor not found")
	}

	return nil
}

// DeleteSensor deletes a sensor by its ID
func (db *DB) DeleteSensor(ctx context.Context, id int) error {
	query := `
		DELETE FROM sensors
		WHERE id = $1
	`

	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting sensor: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("sensor not found")
	}

	return nil
}
