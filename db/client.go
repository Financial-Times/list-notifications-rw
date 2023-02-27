package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	database   string
	collection string
	maxLimit   int
	cacheDelay int
	client     *mongo.Client
	log        *logger.UPPLogger
}

// NewClient creates new client instance
func NewClient(address, database, collection string, cacheDelay, maxLimit int, log *logger.UPPLogger) (*Client, error) {
	uri := fmt.Sprintf("mongodb://%s", address)
	opts := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:     client,
		database:   database,
		collection: collection,
		cacheDelay: cacheDelay,
		maxLimit:   maxLimit,
		log:        log,
	}, nil
}

// WriteNotification inserts a notification into mongo
func (c *Client) WriteNotification(notification *model.InternalNotification) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	collection := c.client.Database(c.database).Collection(c.collection)
	_, err := collection.InsertOne(ctx, notification)
	return err
}

// ReadNotifications reads notifications from the collection.
func (c *Client) ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error) {
	collection := c.client.Database(c.database).Collection(c.collection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	query := generateQuery(c.cacheDelay, offset, c.maxLimit, since, c.log)
	pipe, err := collection.Aggregate(ctx, query)
	if err != nil {
		return nil, err
	}

	var results []model.InternalNotification
	if err = pipe.All(ctx, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

// FindNotificationByTransactionID locates one instance of a notification with the given Transaction ID (publishReference)
func (c *Client) FindNotificationByTransactionID(transactionID string) (model.InternalNotification, error) {
	filter := findByTransactionID(transactionID)
	return c.findNotificationWithFilter(filter)
}

// FindNotificationByPartialTransactionID locates one instance of a notification with the given Transaction ID (publishReference)
func (c *Client) FindNotificationByPartialTransactionID(transactionID string) (model.InternalNotification, error) {
	filter := findByPartialTransactionID(transactionID)
	return c.findNotificationWithFilter(filter)
}

func (c *Client) findNotificationWithFilter(filter bson.M) (model.InternalNotification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	var notification model.InternalNotification
	err := c.
		client.
		Database(c.database).
		Collection(c.collection).
		FindOne(ctx, filter).
		Decode(&notification)
	return notification, err
}

// EnsureIndexes creates indexes
func (c *Client) EnsureIndexes() error {
	collection := c.client.Database(c.database).Collection(c.collection)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	lastModifiedName := "last-modified-index"
	lastModifiedIndex := mongo.IndexModel{
		Keys: bson.M{"lastModified": -1},
		Options: &options.IndexOptions{
			Name: &lastModifiedName,
		},
	}
	publishReferenceName := "publish-reference-index"
	publishReferenceIndex := mongo.IndexModel{
		Keys: bson.M{"publishReference": 1},
		Options: &options.IndexOptions{
			Name: &publishReferenceName,
		},
	}
	uuidName := "uuid-index"
	uuidIndex := mongo.IndexModel{
		Keys: bson.M{"uuid": 1},
		Options: &options.IndexOptions{
			Name: &uuidName,
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{lastModifiedIndex, publishReferenceIndex, uuidIndex})
	return err
}

// GetLimit returns the max number of records returned by a query
func (c *Client) GetLimit() int {
	return c.maxLimit
}

// Ping returns a mongo ping response
func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return c.client.Ping(ctx, nil)
}

// Close closes the entire database connection
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	return c.client.Disconnect(ctx)
}
