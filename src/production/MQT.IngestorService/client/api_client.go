package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements circuit breaker pattern for resilience
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	state        CircuitBreakerState
	failureCount int
	lastFailTime time.Time
	mutex        sync.RWMutex
}

// APIClient handles communication with the API Service
type APIClient struct {
	baseURL        string
	httpClient     *http.Client
	apiSecret      string
	circuitBreaker *CircuitBreaker
	maxRetries     int
	retryDelay     time.Duration
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL, apiSecret string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiSecret: apiSecret,
		circuitBreaker: &CircuitBreaker{
			maxFailures:  5,
			resetTimeout: 30 * time.Second,
			state:        StateClosed,
		},
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
}

// ValidatePiRequest represents the request to validate a Pi
type ValidatePiRequest struct {
	PiID string `json:"pi_id"`
}

// ValidatePiResponse represents the response from Pi validation
type ValidatePiResponse struct {
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

// ValidateDeviceRequest represents the request to validate a Device
type ValidateDeviceRequest struct {
	PiID     string `json:"pi_id"`
	DeviceID int    `json:"device_id"`
}

// ValidateDeviceResponse represents the response from Device validation
type ValidateDeviceResponse struct {
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

// CreateReadingRequest represents the request to create a reading
type CreateReadingRequest struct {
	PiID     string                 `json:"pi_id"`
	DeviceID int                    `json:"device_id"`
	Ts       time.Time              `json:"ts"`
	Payload  map[string]interface{} `json:"payload"`
}

// CreateReadingResponse represents the response from reading creation
type CreateReadingResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Circuit breaker methods
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		return time.Since(cb.lastFailTime) > cb.resetTimeout
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = StateClosed
}

func (cb *CircuitBreaker) onFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.state = StateOpen
	}
}

func (cb *CircuitBreaker) onHalfOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = StateHalfOpen
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (c *APIClient) retryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Check circuit breaker
		if !c.circuitBreaker.canExecute() {
			return fmt.Errorf("circuit breaker is open")
		}

		// Execute operation
		err := operation()
		if err == nil {
			c.circuitBreaker.onSuccess()
			return nil
		}

		lastErr = err
		c.circuitBreaker.onFailure()

		// Don't retry on last attempt
		if attempt == c.maxRetries {
			break
		}

		// Calculate backoff delay
		delay := time.Duration(float64(c.retryDelay) * math.Pow(2, float64(attempt)))

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", c.maxRetries+1, lastErr)
}

// ValidatePi checks if a Pi exists in the API Service
func (c *APIClient) ValidatePi(ctx context.Context, piID string) (bool, error) {
	var result bool
	var resultErr error

	err := c.retryWithBackoff(ctx, func() error {
		req := ValidatePiRequest{PiID: piID}

		resp, err := c.makeRequest(ctx, "POST", "/internal/pis/validate", req)
		if err != nil {
			resultErr = fmt.Errorf("failed to validate Pi: %w", err)
			return resultErr
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			resultErr = fmt.Errorf("API returned status %d", resp.StatusCode)
			return resultErr
		}

		var response ValidatePiResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resultErr = fmt.Errorf("failed to decode response: %w", err)
			return resultErr
		}

		if response.Error != "" {
			resultErr = fmt.Errorf("API error: %s", response.Error)
			return resultErr
		}

		result = response.Exists
		return nil
	})

	if err != nil {
		return false, err
	}

	return result, nil
}

// ValidateDevice checks if a Device exists for a given Pi
func (c *APIClient) ValidateDevice(ctx context.Context, piID string, deviceID int) (bool, error) {
	var result bool
	var resultErr error

	err := c.retryWithBackoff(ctx, func() error {
		req := ValidateDeviceRequest{
			PiID:     piID,
			DeviceID: deviceID,
		}

		resp, err := c.makeRequest(ctx, "POST", "/internal/devices/validate", req)
		if err != nil {
			resultErr = fmt.Errorf("failed to validate Device: %w", err)
			return resultErr
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			resultErr = fmt.Errorf("API returned status %d", resp.StatusCode)
			return resultErr
		}

		var response ValidateDeviceResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resultErr = fmt.Errorf("failed to decode response: %w", err)
			return resultErr
		}

		if response.Error != "" {
			resultErr = fmt.Errorf("API error: %s", response.Error)
			return resultErr
		}

		result = response.Exists
		return nil
	})

	if err != nil {
		return false, err
	}

	return result, nil
}

// CreateReading creates a reading in the API Service
func (c *APIClient) CreateReading(ctx context.Context, reading hardware_models.Reading) error {
	var resultErr error

	err := c.retryWithBackoff(ctx, func() error {
		req := CreateReadingRequest{
			PiID:     reading.PiID,
			DeviceID: reading.DeviceID,
			Ts:       reading.Ts,
			Payload:  reading.Payload,
		}

		resp, err := c.makeRequest(ctx, "POST", "/internal/readings", req)
		if err != nil {
			resultErr = fmt.Errorf("failed to create reading: %w", err)
			return resultErr
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resultErr = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
			return resultErr
		}

		var response CreateReadingResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resultErr = fmt.Errorf("failed to decode response: %w", err)
			return resultErr
		}

		if !response.Success && response.Error != "" {
			resultErr = fmt.Errorf("API error: %s", response.Error)
			return resultErr
		}

		return nil
	})

	return err
}

// makeRequest makes an HTTP request to the API Service
func (c *APIClient) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add service-to-service authentication
	req.Header.Set("Authorization", "Bearer "+c.apiSecret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mqtt-ingestor-service")

	return c.httpClient.Do(req)
}

// Health checks if the API Service is healthy
func (c *APIClient) Health(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, "GET", "/health/live", nil)
	if err != nil {
		return fmt.Errorf("failed to check API health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetCircuitBreakerStatus returns the current circuit breaker status for monitoring
func (c *APIClient) GetCircuitBreakerStatus() map[string]interface{} {
	c.circuitBreaker.mutex.RLock()
	defer c.circuitBreaker.mutex.RUnlock()

	stateStr := "unknown"
	switch c.circuitBreaker.state {
	case StateClosed:
		stateStr = "closed"
	case StateOpen:
		stateStr = "open"
	case StateHalfOpen:
		stateStr = "half-open"
	}

	return map[string]interface{}{
		"state":          stateStr,
		"failure_count":  c.circuitBreaker.failureCount,
		"last_fail_time": c.circuitBreaker.lastFailTime,
		"max_failures":   c.circuitBreaker.maxFailures,
		"reset_timeout":  c.circuitBreaker.resetTimeout,
	}
}
