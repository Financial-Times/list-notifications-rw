package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/list-notifications-rw/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database interface {
	WriteNotification(notification *model.InternalNotification) error
	ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error)
	FindNotificationByTransactionID(transactionID string) (model.InternalNotification, error)
	FindNotificationByPartialTransactionID(transactionID string) (model.InternalNotification, error)
	EnsureIndexes() error
	GetLimit() int
	Ping() error
	Close() error
}

type Client struct {
	database   string
	collection string
	maxLimit   int
	cacheDelay int
	ctx        context.Context
	client     *mongo.Client
	log        *logger.UPPLogger
}

// NewClient creates new client instance
func NewClient(ctx context.Context, address, database, collection string, cacheDelay, maxLimit int, log *logger.UPPLogger) (*Client, error) {
	uri := fmt.Sprintf("mongodb://%s", address)
	opts := options.Client().ApplyURI(uri)

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
	collection := c.client.Database(c.database).Collection(c.collection)
	_, err := collection.InsertOne(c.ctx, notification)
	return err
}

// ReadNotifications reads notifications from the collection.
func (c *Client) ReadNotifications(offset int, since time.Time) (*[]model.InternalNotification, error) {
	collection := c.client.Database(c.database).Collection(c.collection)

	query := generateQuery(c.cacheDelay, offset, c.maxLimit, since)
	pipe, err := collection.Aggregate(c.ctx, query)

	var results []model.InternalNotification
	if err = pipe.All(c.ctx, &results); err != nil {
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
	var notification model.InternalNotification
	err := c.
		client.
		Database(c.database).
		Collection(c.collection).
		FindOne(c.ctx, filter).
		Decode(&notification)
	return notification, err
}

// EnsureIndexes creates indexes
func (c *Client) EnsureIndexes() error {
	collection := c.client.Database(c.database).Collection(c.collection)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	uuidName := "uuid-index"
	uuidIndex := mongo.IndexModel{
		Keys: bson.D{
			primitive.E{Key: "uuid", Value: 1},
		},
		Options: &options.IndexOptions{
			Name: &uuidName,
		},
	}
	publishReferenceName := "publish-reference-index"
	publishReferenceIndex := mongo.IndexModel{
		Keys: bson.D{
			primitive.E{Key: "publishReference", Value: 1},
		},
		Options: &options.IndexOptions{
			Name: &publishReferenceName,
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{uuidIndex, publishReferenceIndex})
	return err
}

// GetLimit returns the max number of records returned by a query
func (c *Client) GetLimit() int {
	return c.maxLimit
}

// Ping returns a mongo ping response
func (c *Client) Ping() error {
	return c.client.Ping(c.ctx, nil)
}

// Close closes the entire database connection
func (c *Client) Close() error {
	return c.client.Disconnect(c.ctx)
}
