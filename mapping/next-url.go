package mapping

import (
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/Financial-Times/list-notifications-rw/model"
)

// NextLinkGenerator returns the link to the next result set in a paginated response.
type NextLinkGenerator interface {
	NextLink(since time.Time, offset int, notifications []model.InternalNotification) model.Link
	ProcessRequestLink(uri *url.URL) *url.URL
}

// OffsetNextLink is the default implementation for NextLinkGenerator
type OffsetNextLink struct {
	ApiHost    string
	CacheDelay int
	MaxLimit   int
}

// NextLink given the since date and offset from the original incoming request, and the notifications read from the db, return the 'next' link
func (o OffsetNextLink) NextLink(since time.Time, offset int, notifications []model.InternalNotification) model.Link {
	updatedSince := o.calculateSince(notifications, since)
	updatedOffset := o.calculateOffset(notifications, offset)

	return o.generateLink(updatedSince, updatedOffset)
}

func (o OffsetNextLink) ProcessRequestLink(uri *url.URL) *url.URL {
	uri.Scheme = "http"
	uri.Host = o.ApiHost
	return uri
}

func (o OffsetNextLink) generateLink(since time.Time, offset int) model.Link {
	uri := url.URL{}
	uri.Scheme = "http"
	uri.Host = o.ApiHost
	uri.Path = "/lists/notifications"
	params := uri.Query()

	if offset > 0 {
		params.Add("offset", strconv.Itoa(offset))
	}

	params.Add("since", since.Format(time.RFC3339Nano))
	uri.RawQuery = params.Encode()

	return model.Link{
		Href: uri.String(),
		Rel:  "next",
	}
}

func (o OffsetNextLink) calculateSince(notifications []model.InternalNotification, since time.Time) time.Time {
	if len(notifications) == 0 {
		return since
	}

	last := notifications[min(len(notifications), o.MaxLimit)-1] // make sure we know this is the last page
	return o.normalizeForCacheDelay(last.LastModified)
}

func (o OffsetNextLink) normalizeForCacheDelay(since time.Time) time.Time {
	return since.Add(time.Duration(o.CacheDelay) * time.Second) // add the cache delay back on to the since date; this would be removed on the next request as normal
}

func (o OffsetNextLink) calculateOffset(notifications []model.InternalNotification, currentOffset int) int {
	if len(notifications) == 0 { // if there are no results
		return currentOffset
	}

	if len(notifications) <= o.MaxLimit { // reset offset if we've arrived at the last page
		return 0
	}

	doBoundaryNotificationsHaveTheSameModifiedDate := checkBoundary(notifications)
	if !doBoundaryNotificationsHaveTheSameModifiedDate { // reset offset if our boundary has different modified dates
		return 0
	}

	size := sizeOfBoundary(notifications)
	if size == len(notifications) { // if every record in this result set has the same date, then increase the initial offset
		size += currentOffset
	}

	return size - 1 // otherwise, return new offset minus one for the extra boundary record
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func checkBoundary(notifications []model.InternalNotification) bool {
	size := len(notifications)
	lastResult, boundaryResult := notifications[size-2], notifications[size-1]

	return lastResult.LastModified == boundaryResult.LastModified
}

func sizeOfBoundary(notifications []model.InternalNotification) int {
	reverse := make([]model.InternalNotification, len(notifications))
	copy(reverse, notifications)

	sort.Sort(sort.Reverse(ByLastModified(reverse)))

	boundaryModifiedDate := reverse[0].LastModified

	size := 0

	for i := range reverse {
		if reverse[i].LastModified == boundaryModifiedDate {
			size++
		} else {
			break
		}
	}

	return size
}

// ByLastModified sorts InternalNotifications by last modified date.
type ByLastModified []model.InternalNotification

func (s ByLastModified) Len() int {
	return len(s)
}

func (s ByLastModified) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByLastModified) Less(i, j int) bool {
	return s[i].LastModified.Before(s[j].LastModified)
}
