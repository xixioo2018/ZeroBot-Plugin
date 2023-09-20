package mongo

import (
	"context"
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/database"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	dbName      string
)

func InitMongo(database2 database.Config) {
	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority",
		database2.Mongo.UserName,
		database2.Mongo.Password,
		database2.Mongo.Hostname,
		database2.Mongo.Database,
	)
	if database2.Mongo.Port != 0 {
		uri = fmt.Sprintf(
			"mongodb://%s:%s@%s:%d/?directConnection=true&retryWrites=true&w=majority",
			database2.Mongo.UserName,
			database2.Mongo.Password,
			database2.Mongo.Hostname,
			database2.Mongo.Port,
			//database2.Mongo.Database,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	mongoClient = client
	dbName = database2.Mongo.Database
}

func Test() {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//utils.PanicNotNil(mongoClient.Ping(ctx, nil))
}

func Collection(collectionName string) *mongo.Collection {
	if mongoClient == nil {
		InitMongo(database.DefaultConfig)
	}
	return mongoClient.Database(dbName).Collection(collectionName)
}
