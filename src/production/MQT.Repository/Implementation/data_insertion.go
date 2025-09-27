package implementation

import (
	"context"
	"time"

	mqtmodels "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Models"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoReadingRepository struct {
	coll *mongo.Collection
}

func NewMongoReadingRepository(coll *mongo.Collection) *MongoReadingRepository {
	return &MongoReadingRepository{coll: coll}
}

func (r *MongoReadingRepository) InsertOne(ctx context.Context, rd mqtmodels.Reading) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := r.coll.InsertOne(ctx, rd)
	return err
}

func (r *MongoReadingRepository) InsertMany(ctx context.Context, rs []mqtmodels.Reading) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	docs := make([]interface{}, 0, len(rs))
	for i := range rs {
		docs = append(docs, rs[i])
	}
	_, err := r.coll.InsertMany(ctx, docs)
	return err
}
