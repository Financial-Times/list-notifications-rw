package mapping

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"

	"github.com/Financial-Times/list-notifications-rw/model"
)

var isUUID = regexp.MustCompile("[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}")

// NotificationsMapper maps notifications from json to internal and internal to public.
type NotificationsMapper interface {
	MapRequestToInternalNotification(uuid string, decoder *json.Decoder) (*model.InternalNotification, error)
	MapInternalNotificationToPublic(notification model.InternalNotification) model.PublicNotification
}

// DefaultMapper is the standard NotificationsMapper implementation
type DefaultMapper struct {
	ApiHost string
}

// MapRequestToInternalNotification maps json (from a decoder) to an InternalNotification
func (m DefaultMapper) MapRequestToInternalNotification(uuid string, decoder *json.Decoder) (*model.InternalNotification, error) {
	notification := &model.InternalNotification{}

	err := decoder.Decode(notification)
	if err != nil {
		return nil, errors.New("Failed to parse json for list body!")
	}

	if !isUUID.MatchString(notification.UUID) {
		return nil, errors.New("List document contained an invalid UUID!")
	}

	if uuid != notification.UUID {
		return nil, errors.New("List document contained a different UUID to the request URI!")
	}

	if notification.EventType == "" {
		notification.EventType = "UPDATE"
	}

	return notification, nil
}

// MapInternalNotificationToPublic maps an InternalNotification to a PublicNotification
func (m DefaultMapper) MapInternalNotificationToPublic(notification model.InternalNotification) model.PublicNotification {
	return model.PublicNotification{
		ID:               m.buildId(notification.UUID),
		APIURL:           m.buildApiUrl(notification.UUID),
		Type:             "http://www.ft.com/thing/ThingChangeType/UPDATE",
		Title:            notification.Title,
		PublishReference: notification.PublishReference,
		LastModified:     notification.LastModified.UTC(),
	}
}

func (m DefaultMapper) buildId(uuid string) string {
	uri, _ := url.Parse("http://" + m.ApiHost + "/things/" + uuid)
	return uri.String()
}

func (m DefaultMapper) buildApiUrl(uuid string) string {
	uri, _ := url.Parse("http://" + m.ApiHost + "/lists/" + uuid)
	return uri.String()
}
