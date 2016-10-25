package model

import (
	"time"
)

// InternalNotification represents the document format within Mongo
type InternalNotification struct {
	Title string `json:"title"`
	UUID string `json:"uuid"`
	EventType string `json:"eventType"`
	PublishReference string `json:"publishReference"`
	ChangeDate time.Time `json:"lastModified,omitempty"`
}

// PublicNotification represents the public format for a notification (seen on read)
type PublicNotification struct {
	Type string `json:"type"`
	Id string `json:"id"`
	ApiUrl string `json:"apiUrl"`
	Title string `json:"title"`
	PublishReference string `json:"publishReference,omitempty"`
	LastModified time.Time `json:"lastModified,omitempty"`
}
