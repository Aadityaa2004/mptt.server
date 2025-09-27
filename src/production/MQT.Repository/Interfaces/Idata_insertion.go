package interfaces

import (
	"context"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
)

type ReadingRepository interface {
	InsertOne(ctx context.Context, r mqtmodels.Reading) error
	InsertMany(ctx context.Context, rs []mqtmodels.Reading) error
}
