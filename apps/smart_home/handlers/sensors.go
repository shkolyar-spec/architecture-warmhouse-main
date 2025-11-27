package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"smarthome/db"
	"smarthome/models"
	"smarthome/services"

	"github.com/gin-gonic/gin"
)

// SensorHandler handles sensor-related requests
type SensorHandler struct {
	DB                 *db.DB
	TemperatureService *services.TemperatureService
}

// NewSensorHandler creates a new SensorHandler
func NewSensorHandler(db *db.DB, temperatureService *services.TemperatureService) *SensorHandler {
	return &SensorHandler{
		DB:                 db,
		TemperatureService: temperatureService,
	}
}

// RegisterRoutes registers sensor routes
func (h *SensorHandler) RegisterRoutes(rg *gin.RouterGroup) {
	sensors := rg.Group("/sensors")
	{
		sensors.GET("", h.GetSensors)
		sensors.GET("/:id", h.GetSensorByID)
		sensors.POST("", h.CreateSensor)
		sensors.PUT("/:id", h.UpdateSensor)
		sensors.PATCH("/:id/value", h.UpdateSensorValue)
		sensors.DELETE("/:id", h.DeleteSensor)
	}
}

// GetSensors handles GET /api/v1/sensors
func (h *SensorHandler) GetSensors(c *gin.Context) {
	sensors, err := h.DB.GetSensors(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich temperature sensors with random values from temperature-api
	for i := range sensors {
		if sensors[i].Type == models.Temperature {
			temp, err := h.TemperatureService.GetTemperature(sensors[i].Location)
			if err != nil {
				log.Printf("failed to get temperature for %s: %v", sensors[i].Location, err)
				continue
			}
			sensors[i].Value = temp.Value
			sensors[i].Unit = temp.Unit
			sensors[i].LastUpdated = temp.Timestamp
		}
	}

	c.JSON(http.StatusOK, sensors)
}

// GetSensorByID handles GET /api/v1/sensors/:id
func (h *SensorHandler) GetSensorByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor ID"})
		return
	}

	sensor, err := h.DB.GetSensorByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sensor not found"})
		return
	}

	if sensor.Type == models.Temperature {
		temp, err := h.TemperatureService.GetTemperature(sensor.Location)
		if err == nil && temp != nil {
			sensor.Value = temp.Value
			sensor.Unit = temp.Unit
			sensor.LastUpdated = temp.Timestamp
		}
	}

	c.JSON(http.StatusOK, sensor)
}

// CreateSensor handles POST /api/v1/sensors
func (h *SensorHandler) CreateSensor(c *gin.Context) {
	var req models.SensorCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sensor, err := h.DB.CreateSensor(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sensor)
}

// UpdateSensor handles PUT /api/v1/sensors/:id
func (h *SensorHandler) UpdateSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor ID"})
		return
	}

	var req models.SensorUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.DB.UpdateSensor(context.Background(), id, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("sensor %d not found", id)})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// UpdateSensorValue handles PATCH /api/v1/sensors/:id/value
func (h *SensorHandler) UpdateSensorValue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor ID"})
		return
	}

	var request struct {
		Value  float64 `json:"value" binding:"required"`
		Status string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := request.Status
	if status == "" {
		status = "active"
	}

	err = h.DB.UpdateSensorValue(context.Background(), id, request.Value, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sensor value updated successfully"})
}

// DeleteSensor handles DELETE /api/v1/sensors/:id
func (h *SensorHandler) DeleteSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sensor ID"})
		return
	}

	if err := h.DB.DeleteSensor(context.Background(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
