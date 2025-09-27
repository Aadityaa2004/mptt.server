package mqtingestor

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	interfaces "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Repository/Interfaces"
)

type Ingestor struct {
	cfg    mqtmodels.IngestorConfig
	repo   interfaces.ReadingRepository
	client mqtt.Client
	msgCh  chan mqtmodels.Reading
	wg     sync.WaitGroup
}

func New(cfg mqtmodels.IngestorConfig, r interfaces.ReadingRepository) *Ingestor {
	return &Ingestor{
		cfg:   cfg,
		repo:  r,
		msgCh: make(chan mqtmodels.Reading, 4096),
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
		log.Printf("mqtt lost: %v", err)
	}
	opts.OnConnect = func(c mqtt.Client) {
		topic := i.cfg.Topic
		if i.cfg.SharedGroup != "" {
			topic = fmt.Sprintf("$share/%s/%s", i.cfg.SharedGroup, i.cfg.Topic)
		}
		log.Printf("mqtt connected, subscribing to %s", topic)
		if token := c.Subscribe(topic, 1, i.onMessage); token.Wait() && token.Error() != nil {
			log.Printf("subscribe error: %v", token.Error())
		}
	}

	i.client = mqtt.NewClient(opts)
	if tk := i.client.Connect(); tk.Wait() && tk.Error() != nil {
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
	if i.client != nil && i.client.IsConnected() {
		i.client.Disconnect(500)
	}
	close(i.msgCh)
	i.wg.Wait()
}

func (i *Ingestor) IsConnected() bool {
	return i.client != nil && i.client.IsConnected()
}

func (i *Ingestor) onMessage(_ mqtt.Client, m mqtt.Message) {
	log.Printf("Received MQTT message on topic: %s, payload: %s", m.Topic(), string(m.Payload()))

	var payload map[string]interface{}
	if err := json.Unmarshal(m.Payload(), &payload); err != nil {
		payload = map[string]interface{}{"raw": string(m.Payload())}
	}

	dev := ""
	parts := strings.Split(m.Topic(), "/")
	if len(parts) >= 2 {
		dev = parts[1] // e.g., sensors/<deviceId>/metric
	}

	reading := mqtmodels.Reading{
		Topic:      m.Topic(),
		DeviceID:   dev,
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}

	log.Printf("Queuing reading for device: %s", dev)
	i.msgCh <- reading
}

func (i *Ingestor) batchWriter(ctx context.Context) {
	batch := make([]mqtmodels.Reading, 0, i.cfg.BatchSize)
	timer := time.NewTimer(i.cfg.BatchWindow)
	defer timer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		log.Printf("Flushing batch of %d readings to MongoDB", len(batch))
		if len(batch) == 1 {
			if err := i.repo.InsertOne(ctx, batch[0]); err != nil {
				log.Printf("Error inserting reading: %v", err)
			} else {
				log.Printf("Successfully inserted 1 reading")
			}
		} else {
			if err := i.repo.InsertMany(ctx, batch); err != nil {
				log.Printf("Error inserting readings: %v", err)
			} else {
				log.Printf("Successfully inserted %d readings", len(batch))
			}
		}
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
