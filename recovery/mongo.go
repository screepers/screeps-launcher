package recovery

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoData map[string][]bson.Raw

func (r *Recovery) getMongoClient(ctx context.Context) (*mongo.Client, *mongo.Database, error) {
	host := "localhost"
	port := "27017"
	database := "screeps"
	if v, ok := r.config.Env.Shared["MONGO_HOST"]; ok {
		host = v
	}
	if v, ok := r.config.Env.Shared["MONGO_PORT"]; ok {
		port = v
	}
	if v, ok := r.config.Env.Shared["MONGO_DATABASE"]; ok {
		database = v
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port)))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to connect")
	}
	db := client.Database(database)
	return client, db, nil
}

func (r *Recovery) mongoBackup() (mongoData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client, db, err := r.getMongoClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get client")
	}
	defer client.Disconnect(ctx)
	data := mongoData{}
	cols, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list collections")
	}
	for _, name := range cols {
		cur, err := db.Collection(name).Find(ctx, bson.M{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get records from collection %s", name)
		}
		arr := make([]bson.Raw, 0)
		defer cur.Close(ctx)
		for cur.Next(ctx) {
			arr = append(arr, cur.Current)
		}
		if len(arr) > 0 {
			data[name] = arr
		}
	}
	return data, nil
}

func (r *Recovery) mongoRestore(data mongoData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	client, db, err := r.getMongoClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get client")
	}
	defer client.Disconnect(ctx)
	err = db.Drop(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to drop db")
	}
	for name, docs := range data {
		data := make([]interface{}, len(docs))
		for i, d := range docs {
			data[i] = d
		}
		_, err := db.Collection(name).InsertMany(ctx, data)
		if err != nil {
			return errors.Wrapf(err, "failed to insert into collection %s", name)
		}
	}
	return nil
}
