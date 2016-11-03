package model

import (
	"time"
)

// InternalNotification represents the document format within Mongo
type InternalNotification struct {
	Title            string    `json:"title" bson:"title"`
	UUID             string    `json:"uuid" bson:"uuid"`
	EventType        string    `json:"eventType" bson:"eventType"`
	PublishReference string    `json:"publishReference" bson:"publishReference"`
	LastModified     time.Time `json:"lastModified,omitempty" bson:"lastModified,omitempty"`
}

// PublicNotification represents the public format for a notification (seen on read)
type PublicNotification struct {
	Type             string    `json:"type"`
	ID               string    `json:"id"`
	APIURL           string    `json:"apiUrl"`
	Title            string    `json:"title"`
	PublishReference string    `json:"publishReference,omitempty"`
	LastModified     time.Time `json:"lastModified,omitempty"`
}

// Link represents the next url in the notification page
type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// PublicNotificationPage represents one page in the result set
type PublicNotificationPage struct {
	RequestURL    string               `json:"requestUrl"`
	Notifications []PublicNotification `json:"notifications"`
	Links         []Link               `json:"links"`
}
