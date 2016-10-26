package mapping

import (
	"testing"
	"github.com/Financial-Times/list-notifications-rw/model"
	"time"
	"github.com/stretchr/testify/assert"
	"net/url"
)

var nextLink = OffsetNextLink{ApiHost: "go-tests.ft.com", MaxLimit: 200, CacheDelay: 10}


func TestNextLink(t *testing.T){
	now := time.Now()
	since := now.Add(-20 * time.Second)

	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now.Add(-5 * time.Second)},
		{LastModified: now.Add(-3 * time.Second)},
		{LastModified: now.Add(-1 * time.Second)},
		{LastModified: now},
		{LastModified: now},
		{LastModified: now},
	}

	nextLink.MaxLimit = 6
	nextLink.CacheDelay = 5

	// We test each part individually, so use these to create expected results
	calculated := nextLink.calculateSince(notifications, since)
	offset := nextLink.calculateOffset(notifications, 10)

	link := nextLink.NextLink(since, 10, notifications)
	assert.Equal(t, "next", link.Rel, "Should be hardcoded to next.")
	assert.Equal(t, nextLink.generateLink(calculated, offset).Href, link.Href, "Should match generated link.")
}

func TestGenerateLinkWithOffset(t *testing.T){
	now := time.Now().UTC()

	link := nextLink.generateLink(now, 10)
	uri, _ := url.Parse("http://go-tests.ft.com/lists/notifications?since=" + now.Format(time.RFC3339Nano) + "&offset=10")
	uri.RawQuery = uri.Query().Encode()

	assert.Equal(t, uri.String(), link.Href, "This is the link we should generate")
}

func TestGenerateLinkWithoutOffset(t *testing.T){
	now := time.Now().UTC()

	link := nextLink.generateLink(now, 0)
	uri, _ := url.Parse("http://go-tests.ft.com/lists/notifications?since=" + now.Format(time.RFC3339Nano))
	uri.RawQuery = uri.Query().Encode()

	assert.Equal(t, uri.String(), link.Href, "This is the link we should generate")
}

func TestBoundaryHasSameDate(t *testing.T){
	now := time.Now()

	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now},
		{LastModified: now},
	}

	result := checkBoundary(notifications)
	assert.True(t, result, "Boundary has the same change date, so should be true.")
}

func TestBoundaryHasDifferentDate(t *testing.T){
	now := time.Now()

	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now.Add(-8 * time.Second)},
		{LastModified: now},
	}

	result := checkBoundary(notifications)
	assert.False(t, result, "Boundary has a different change date, so should be false.")
}

func TestBoundarySize(t *testing.T){
	now := time.Now()

	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now.Add(-6 * time.Second)},
		{LastModified: now},
		{LastModified: now},
		{LastModified: now},
	}

	result := sizeOfBoundary(notifications)
	assert.Equal(t, 3, result, "Boundary should be 3.")
	assert.False(t, notifications[0].LastModified == now, "The collection should remain the same!")
}

func TestCalculateOffsetNoNotifications(t *testing.T){
	notifications := make([]model.InternalNotification, 0)
	offset := nextLink.calculateOffset(notifications, 0)
	assert.Equal(t, 0, offset, "Should equal current offset (which is 0)")
}

func TestCalculateOffsetLessThanMax(t *testing.T){
	notifications := make([]model.InternalNotification, 2)
	offset := nextLink.calculateOffset(notifications, 80)
	assert.Equal(t, 0, offset, "Should equal 0, regardless of current offset")
}

func TestCalculateOffsetNoBoundary(t *testing.T){
	now := time.Now()

	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now.Add(-5 * time.Second)},
		{LastModified: now},
	}

	nextLink.MaxLimit = 2
	offset := nextLink.calculateOffset(notifications, 80)
	assert.Equal(t, 0, offset, "Should equal 0, regardless of current offset")
}

func TestCalculateOffsetAllSame(t *testing.T){
	now := time.Now()

	notifications := []model.InternalNotification{
		{LastModified: now},
		{LastModified: now},
		{LastModified: now},
	}

	nextLink.MaxLimit = 2
	offset := nextLink.calculateOffset(notifications, 5)
	assert.Equal(t, 7, offset, "Should equal current offset + size of collection - 1 = 7")
}

func TestCalculateOffset(t *testing.T){
	now := time.Now()

	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now.Add(-5 * time.Second)},
		{LastModified: now},
		{LastModified: now},
	}

	nextLink.MaxLimit = 3
	offset := nextLink.calculateOffset(notifications, 5)
	assert.Equal(t, 1, offset, "Should equal size of boundary (2) - 1 = 1")
}

func TestCalculateSinceNoResults(t *testing.T){
	now := time.Now()

	notifications := make([]model.InternalNotification, 0)
	since := nextLink.calculateSince(notifications, now)
	assert.Equal(t, now, since, "Length 0, so should return the same since date")
}

func TestCalculateSinceFullPage(t *testing.T){
	now := time.Now()
	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now.Add(-5 * time.Second)},
		{LastModified: now.Add(-3 * time.Second)},
		{LastModified: now.Add(-1 * time.Second)},
	}

	nextLink.MaxLimit = 3
	nextLink.CacheDelay = 5
	since := nextLink.calculateSince(notifications, now.Add(-20 * time.Second))
	assert.Equal(t, now.Add(-3 * time.Second).Add(5 * time.Second), since, "Should return second to last in set + cache delay")
}

func TestCalculateSincePartialPage(t *testing.T){
	now := time.Now()
	notifications := []model.InternalNotification{
		{LastModified: now.Add(-10 * time.Second)},
		{LastModified: now.Add(-5 * time.Second)},
		{LastModified: now.Add(-3 * time.Second)},
		{LastModified: now.Add(-1 * time.Second)},
	}

	nextLink.MaxLimit = 6
	nextLink.CacheDelay = 5
	since := nextLink.calculateSince(notifications, now.Add(-20 * time.Second))
	assert.Equal(t, now.Add(-1 * time.Second).Add(5 * time.Second), since, "Should return last in set + cache delay.")
}