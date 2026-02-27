package main

import (
	"context"
	"fmt"

	container "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Container"
	"gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/client"
	mqtingestor "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.IngestorService/ingestor"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
)

// InitializeIngestorApp wires all MQTT Ingestor dependencies and returns
// a configured Ingestor instance, config, API client, logger and shutdown function.
func InitializeIngestorApp() (*mqtingestor.Ingestor, *config.IngestorConfig, client.APIReadingsClient, *logger.Logger, func(context.Context) error, error) {
	ctr, err := container.NewIngestorContainer()
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to initialize ingestor container: %w", err)
	}

	cfg := ctr.GetConfig()
	log := ctr.GetLogger()
	log.Info("Initializing MQTT Ingestor Service")

	apiClient := client.NewAPIClient(cfg.ApiServiceURL, cfg.InternalAPISecret)
	ingestorCfg := mqtingestor.LoadFromEnv()

	ing := mqtingestor.New(ingestorCfg, apiClient, log)

	shutdown := func(ctx context.Context) error {
		ing.Stop()
		return ctr.Shutdown(ctx)
	}

	return ing, cfg, apiClient, log, shutdown, nil
}

