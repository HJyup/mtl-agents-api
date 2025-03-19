package consul

import (
	"context"
	"errors"
	consul "github.com/hashicorp/consul/api"
	"log"
	"strconv"
	"strings"
)

type Registry struct {
	client *consul.Client
}

func NewRegistry(addr string) (*Registry, error) {
	config := consul.DefaultConfig()
	config.Address = addr

	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Registry{client: client}, nil
}

func (r Registry) Register(instanceID, serverName, hostPort string) error {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return errors.New("invalid hostPort")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.New("invalid port")
	}
	host := parts[0]

	return r.client.Agent().ServiceRegister(&consul.AgentServiceRegistration{
		ID:      instanceID,
		Address: host,
		Port:    port,
		Name:    serverName,
		Check: &consul.AgentServiceCheck{
			CheckID:                        instanceID,
			TLSSkipVerify:                  true,
			TTL:                            "5s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "10s",
		},
	})
}

func (r Registry) DeRegister(instanceID string) error {
	log.Println("DeRegistering service with ID:", instanceID)
	return r.client.Agent().ServiceDeregister(instanceID)
}

func (r Registry) Discover(_ context.Context, serverName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serverName, "", true, nil)
	if err != nil {
		return nil, err
	}

	var instances []string
	for _, entry := range entries {
		instances = append(instances, entry.Service.Address+":"+strconv.Itoa(entry.Service.Port))
	}

	return instances, nil
}

func (r Registry) HealthCheck(instanceID string) error {
	return r.client.Agent().UpdateTTL(instanceID, "online", "pass")
}
