package database

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DBservice struct {
	Client *mongo.Client
	DB     *mongo.Database
}

const dbName = "go-redis"
const mongoURI = "mongodb://localhost:27017/" + dbName

func NewDbService() (*DBservice, error) {

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	return &DBservice{
		Client: client,
		DB:     client.Database(dbName),
	}, nil
}

// var DBconn *DBservice

// func InitDB() error {
// 	var err error

// 	DBconn, err = NewDbService()

// 	if err != nil {
// 		return err
// 	}

// 	collections := []string{"users", "chat"}
// 	// Ensure collection creation only if DB is successfully initialized
// 	for _, col := range collections {
// 		// DBconn.DB.CreateCollection(context.TODO(), col)
// 		err = DBconn.DB.CreateCollection(context.TODO(), col)
// 		if err != nil && !mongo.IsDuplicateKeyError(err) {
// 			return err
// 		}
// 	}

// 	return nil
// }
