package mqtmodels

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reading struct {
	ID         primitive.ObjectID      `bson:"_id,omitempty" json:"id,omitempty"`
	Topic      string                  `bson:"topic" json:"topic"`
	DeviceID   string                  `bson:"device_id,omitempty" json:"device_id,omitempty"`
	Payload    map[string]interface{}  `bson:"payload" json:"payload"`
	ReceivedAt time.Time               `bson:"received_at" json:"received_at"`
}