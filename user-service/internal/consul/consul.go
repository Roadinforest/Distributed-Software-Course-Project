package consul

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulClient Consul客户端
type ConsulClient struct {
	client  *api.Client
	agent   *api.Agent
	kv      *api.KV
	cfg     *api.Config
	address string
}

// NewConsulClient 创建Consul客户端
func NewConsulClient(address string) (*ConsulClient, error) {
	cfg := api.DefaultConfig()
	cfg.Address = address
	cfg.WaitTime = 10 * time.Second

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create consul client failed: %w", err)
	}

	return &ConsulClient{
		client:  client,
		agent:   client.Agent(),
		kv:      client.KV(),
		cfg:     cfg,
		address: address,
	}, nil
}

// RegisterService 注册服务到Consul
func (c *ConsulClient) RegisterService(name string, port int, tags []string, healthCheckPath string) error {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = fmt.Sprintf("%s-%d", name, port)
	}

	// 健康检查配置
	check := &api.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://localhost:%d%s", port, healthCheckPath),
		Interval:                       "10s",
		Timeout:                        "5s",
		DeregisterCriticalServiceAfter: "30s",
	}

	service := &api.AgentServiceRegistration{
		ID:      instanceID,
		Name:    name,
		Port:    port,
		Tags:    tags,
		Address: "localhost",
		Check:   check,
	}

	if err := c.agent.ServiceRegister(service); err != nil {
		return fmt.Errorf("register service failed: %w", err)
	}

	log.Printf("service registered: id=%s name=%s port=%d", instanceID, name, port)
	return nil
}

// DeregisterService 从Consul注销服务
func (c *ConsulClient) DeregisterService(instanceID string) error {
	if err := c.agent.ServiceDeregister(instanceID); err != nil {
		return fmt.Errorf("deregister service failed: %w", err)
	}
	log.Printf("service deregistered: id=%s", instanceID)
	return nil
}

// GetService 获取服务实例列表
func (c *ConsulClient) GetService(name string) ([]*api.ServiceEntry, error) {
	services, _, err := c.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("get service failed: %w", err)
	}
	return services, nil
}

// GetConfig 从Consul KV获取配置
func (c *ConsulClient) GetConfig(key string) ([]byte, error) {
	pair, _, err := c.kv.Get(key, nil)
	if err != nil {
		return nil, fmt.Errorf("get config failed: %w", err)
	}
	if pair == nil {
		return nil, fmt.Errorf("config not found: %s", key)
	}
	return pair.Value, nil
}

// PutConfig 写入配置到Consul KV
func (c *ConsulClient) PutConfig(key string, value []byte) error {
	pair := &api.KVPair{
		Key:   key,
		Value: value,
	}
	if _, err := c.kv.Put(pair, nil); err != nil {
		return fmt.Errorf("put config failed: %w", err)
	}
	log.Printf("config saved: key=%s", key)
	return nil
}

// WatchConfig 监听配置变化
func (c *ConsulClient) WatchConfig(key string, callback func([]byte)) error {
	opts := &api.QueryOptions{
		WaitTime: 30 * time.Second,
	}

	for {
		pair, meta, err := c.kv.Get(key, opts)
		if err != nil {
			return fmt.Errorf("watch config failed: %w", err)
		}

		if pair != nil {
			callback(pair.Value)
		}

		opts.WaitIndex = meta.LastIndex
	}
}

// GetDatacenters 获取所有数据中心
func (c *ConsulClient) GetDatacenters() ([]string, error) {
	datacenters, err := c.client.Catalog().Datacenters()
	if err != nil {
		return nil, fmt.Errorf("get datacenters failed: %w", err)
	}
	return datacenters, nil
}

// Close 关闭客户端
func (c *ConsulClient) Close() {
	// Consul Go客户端不需要显式关闭
}
