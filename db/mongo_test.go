package db

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

func startMongo(t *testing.T, limit int) (*MongoDB, func()) {
	if testing.Short() {
		t.Skip("Mongo integration for long tests only.")
	}

	client, err := docker.NewClientFromEnv()

	if err != nil {
		t.Fatalf("Cannot connect to Docker daemon: %s", err)
	}

	c, err := client.CreateContainer(createOptions("mongo"))
	if err != nil {
		t.Fatalf("Cannot create Docker container: %s", err)
	}

	cleanup := func() {
		t.Log("Cleaning up mongo container.")
		if err = client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    c.ID,
			Force: true,
		}); err != nil {
			t.Fatalf("cannot remove container: %s", err)
		}
	}

	err = client.StartContainer(c.ID, &docker.HostConfig{})
	if err != nil {
		t.Fatalf("Cannot start Docker container: %s", err)
	}

	if err = waitStarted(client, c.ID, time.Second*10); err != nil {
		t.Fatalf("Couldn't reach Mongo for testing, aborting.")
	}

	c, err = client.InspectContainer(c.ID)
	if err != nil {
		t.Fatalf("Couldn't inspect container: %s", err)
	}

	var port int64 = 27017
	if len(c.NetworkSettings.PortMappingAPI()) == 0 {
		t.Fatal("No mapped ports!")
	}

	for _, mapping := range c.NetworkSettings.PortMappingAPI() {
		if mapping.PrivatePort == 27017 {
			port = mapping.PublicPort
		}
	}

	dockerURL, err := url.Parse(client.Endpoint())
	if err != nil {
		t.Fatal("Docker host endpoint should be a valid url!")
	}

	host, _, _ := net.SplitHostPort(dockerURL.Host)
	t.Log("Starting Mongo instance at " + host + ":" + strconv.FormatInt(port, 10))

	assert.NoError(t, err, "Should be proper url!")

	mongo := &MongoDB{
		Urls:       host + ":" + strconv.FormatInt(port, 10),
		Timeout:    3800,
		MaxLimit:   limit,
		CacheDelay: 10,
	}

	return mongo, cleanup
}

func waitStarted(client *docker.Client, id string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := client.InspectContainer(id)
		if err != nil {
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start container %s for %v", id, maxWait)
}

func createOptions(dbname string) docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})
	ports["27017/tcp"] = struct{}{}
	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        dbname,
			ExposedPorts: ports,
		},
		HostConfig: &docker.HostConfig{
			PublishAllPorts: true,
		},
	}

	return opts
}
