package mqtingestor

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/client"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	hardware_models "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models/hardware"
)

type Ingestor struct {
	cfg        mqtmodels.IngestorConfig
	apiClient  *client.APIClient
	mqttClient mqtt.Client
	msgCh      chan hardware_models.ReadingWithTopic
	wg         sync.WaitGroup
	logger     *logger.Logger
}

func New(cfg mqtmodels.IngestorConfig, apiClient *client.APIClient, logger *logger.Logger) *Ingestor {
	return &Ingestor{
		cfg:       cfg,
		apiClient: apiClient,
		msgCh:     make(chan hardware_models.ReadingWithTopic, 4096),
		logger:    logger,
	}
}

func (i *Ingestor) Start(ctx context.Context) error {
	opts := mqtt.NewClientOptions().
		AddBroker(i.brokerURL()).
		SetClientID(i.cfg.ClientID).
		SetOrderMatters(false).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetCleanSession(false)

	if i.cfg.BrokerUser != "" {
		opts.SetUsername(i.cfg.BrokerUser)
		opts.SetPassword(i.cfg.BrokerPass)
	}

	if i.cfg.UseTLS {
		tlsCfg, err := i.tlsConfig(i.cfg.CACertPath)
		if err != nil {
			return err
		}
		opts.SetTLSConfig(tlsCfg)
	}

	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		i.logger.Logger.Error().Err(err).Msg("MQTT connection lost")
	}
	opts.OnConnect = func(c mqtt.Client) {
		topic := i.cfg.Topic
		if i.cfg.SharedGroup != "" {
			topic = fmt.Sprintf("$share/%s/%s", i.cfg.SharedGroup, i.cfg.Topic)
		}
		i.logger.Logger.Info().Str("topic", topic).Msg("MQTT connected, subscribing to topic")
		if token := c.Subscribe(topic, 1, i.onMessage); token.Wait() && token.Error() != nil {
			i.logger.Logger.Error().Err(token.Error()).Str("topic", topic).Msg("Failed to subscribe to MQTT topic")
		}
	}

	i.mqttClient = mqtt.NewClient(opts)
	if tk := i.mqttClient.Connect(); tk.Wait() && tk.Error() != nil {
		return tk.Error()
	}

	// batch writer
	i.wg.Add(1)
	go func() {
		defer i.wg.Done()
		i.batchWriter(ctx)
	}()

	return nil
}

func (i *Ingestor) Stop() {
	if i.mqttClient != nil && i.mqttClient.IsConnected() {
		i.mqttClient.Disconnect(500)
	}
	close(i.msgCh)
	i.wg.Wait()
}

func (i *Ingestor) IsConnected() bool {
	return i.mqttClient != nil && i.mqttClient.IsConnected()
}

func (i *Ingestor) onMessage(_ mqtt.Client, m mqtt.Message) {
	i.logger.Logger.Debug().Str("topic", m.Topic()).Str("payload", string(m.Payload())).Msg("Received MQTT message")

	var payload map[string]interface{}
	if err := json.Unmarshal(m.Payload(), &payload); err != nil {
		payload = map[string]interface{}{"raw": string(m.Payload())}
	}

	// Parse topic to extract pi_id and device_id
	// Expected format: sensors/<pi_id>/<device_id>/<metric>
	parts := strings.Split(m.Topic(), "/")
	if len(parts) < 4 {
		i.logger.Logger.Warn().Str("topic", m.Topic()).Str("expected", "sensors/<pi_id>/<device_id>/<metric>").Msg("Invalid topic format")
		// Try to extract pi_id and device_id from what we have for error reporting
		piID := "unknown"
		deviceID := "unknown"
		if len(parts) >= 2 {
			piID = parts[1]
		}
		if len(parts) >= 3 {
			deviceID = parts[2]
		}
		i.publishError(piID, deviceID, "invalid_topic", fmt.Sprintf("Invalid topic format: %s, expected: sensors/<pi_id>/<device_id>/<metric>", m.Topic()))
		return
	}

	piID := parts[1]     // e.g., sensors/pi_001/temperature/humidity -> pi_001
	deviceID := parts[2] // e.g., sensors/pi_001/temperature/humidity -> temperature

	reading := hardware_models.ReadingWithTopic{
		PiID:       piID,
		DeviceID:   deviceID,
		Topic:      m.Topic(),
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}

	i.logger.Logger.Debug().Str("pi_id", piID).Str("device_id", deviceID).Msg("Queuing reading")
	i.msgCh <- reading
}

func (i *Ingestor) batchWriter(ctx context.Context) {
	batch := make([]hardware_models.ReadingWithTopic, 0, i.cfg.BatchSize)
	timer := time.NewTimer(i.cfg.BatchWindow)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		i.logger.Logger.Info().Int("batch_size", len(batch)).Msg("Flushing batch to API Service")

		// Process each reading in the batch
		for _, readingWithTopic := range batch {
			// Convert deviceID string to int
			deviceIDInt, err := strconv.Atoi(readingWithTopic.DeviceID)
			if err != nil {
				i.logger.Logger.Error().Err(err).Str("device_id", readingWithTopic.DeviceID).Msg("Error converting device_id to int")
				continue
			}

			// Validate Pi exists via API
			piExists, err := i.apiClient.ValidatePi(ctx, readingWithTopic.PiID)
			if err != nil {
				i.logger.Logger.Error().Err(err).Str("pi_id", readingWithTopic.PiID).Msg("Failed to validate Pi via API")
				i.publishError(readingWithTopic.PiID, readingWithTopic.DeviceID, "pi_validation_error", fmt.Sprintf("Failed to validate Pi %s: %v", readingWithTopic.PiID, err))
				continue
			}
			if !piExists {
				i.logger.Logger.Warn().Str("pi_id", readingWithTopic.PiID).Msg("Skipping reading: pi not found")
				i.publishError(readingWithTopic.PiID, readingWithTopic.DeviceID, "pi_not_found", fmt.Sprintf("Pi %s does not exist", readingWithTopic.PiID))
				continue
			}

			// Validate device exists via API
			deviceExists, err := i.apiClient.ValidateDevice(ctx, readingWithTopic.PiID, deviceIDInt)
			if err != nil {
				i.logger.Logger.Error().Err(err).Str("pi_id", readingWithTopic.PiID).Int("device_id", deviceIDInt).Msg("Failed to validate Device via API")
				i.publishError(readingWithTopic.PiID, readingWithTopic.DeviceID, "device_validation_error", fmt.Sprintf("Failed to validate Device %d: %v", deviceIDInt, err))
				continue
			}
			if !deviceExists {
				i.logger.Logger.Warn().Str("pi_id", readingWithTopic.PiID).Int("device_id", deviceIDInt).Msg("Skipping reading: device not found")
				i.publishError(readingWithTopic.PiID, readingWithTopic.DeviceID, "device_not_found", fmt.Sprintf("Device %d does not exist for Pi %s", deviceIDInt, readingWithTopic.PiID))
				continue
			}

			// Create reading via API
			reading := hardware_models.Reading{
				PiID:     readingWithTopic.PiID,
				DeviceID: deviceIDInt,
				Ts:       readingWithTopic.ReceivedAt,
				Payload:  readingWithTopic.Payload,
			}
			if err := i.apiClient.CreateReading(ctx, reading); err != nil {
				i.logger.Logger.Error().Err(err).Str("pi_id", readingWithTopic.PiID).Str("device_id", readingWithTopic.DeviceID).Msg("Error creating reading via API")
				i.publishError(readingWithTopic.PiID, readingWithTopic.DeviceID, "create_reading_error", fmt.Sprintf("Failed to create reading: %v", err))
			}
		}

		i.logger.Logger.Info().Int("count", len(batch)).Msg("Successfully processed readings")
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case rd, ok := <-i.msgCh:
			if !ok {
				flush()
				return
			}
			batch = append(batch, rd)
			if len(batch) >= i.cfg.BatchSize {
				flush()
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(i.cfg.BatchWindow)
			}
		case <-timer.C:
			flush()
			timer.Reset(i.cfg.BatchWindow)
		}
	}
}

func (i *Ingestor) brokerURL() string {
	scheme := "tcp"
	if i.cfg.UseTLS {
		scheme = "tcps"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, i.cfg.BrokerHost, i.cfg.BrokerPort)
}

func (i *Ingestor) tlsConfig(caFile string) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if caFile == "" {
		return cfg, nil
	}
	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("bad CA file")
	}
	cfg.RootCAs = cp
	return cfg, nil
}

// publishError publishes an error message to the error topic for Pi feedback
func (i *Ingestor) publishError(piID, deviceID, errorType, message string) {
	if i.mqttClient == nil || !i.mqttClient.IsConnected() {
		return
	}

	errorPayload := map[string]interface{}{
		"error_type": errorType,
		"message":    message,
		"pi_id":      piID,
		"device_id":  deviceID,
		"timestamp":  time.Now().UTC(),
	}

	payloadJSON, err := json.Marshal(errorPayload)
	if err != nil {
		i.logger.Logger.Error().Err(err).Msg("Failed to marshal error payload")
		return
	}

	errorTopic := fmt.Sprintf("ingestor/errors/%s/%s", piID, deviceID)
	token := i.mqttClient.Publish(errorTopic, 1, false, payloadJSON)

	if token.Wait() && token.Error() != nil {
		i.logger.Logger.Error().Err(token.Error()).Str("topic", errorTopic).Msg("Failed to publish error")
	} else {
		i.logger.Logger.Info().Str("topic", errorTopic).Str("message", message).Msg("Published error")
	}
}
