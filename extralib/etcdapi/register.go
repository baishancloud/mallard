package etcdapi

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/client"
)

const (
	// RegisterPrefix is key directory in etcd for registar
	RegisterPrefix = "service_register"
	// RegisterLogPrefix is key directory in etcd for registar history
	RegisterLogPrefix = "service_register_log"
)

// Service is service infomation
type Service struct {
	Name      string            `json:"name,omitempty"`
	Endpoint  string            `json:"endpoint,omitempty"`
	Version   string            `json:"version,omitempty"`
	BuildTime string            `json:"build_time,omitempty"`
	Value     map[string]string `json:"value,omitempty"`
}

// KeyPrefix returns prefix of the service
func (s Service) KeyPrefix() string {
	return fmt.Sprintf("/%s/%s/", s.Name, s.Endpoint)
}

var (
	serviceKey string
)

// Register register service
func Register(svc Service, config interface{}, heartbeat time.Duration) error {
	if cliKv == nil {
		return ErrClientKvNil
	}
	ttl := heartbeat*3 + time.Second
	key, value, err := registerService(svc, ttl)
	if err != nil {
		return err
	}
	if err = registerLog(svc, value); err != nil {
		return err
	}
	if config != nil {
		if err = setConfig(svc, config); err != nil {
			return err
		}
	}
	go keepHeart(key, value, heartbeat, ttl)
	serviceKey = key
	log.Info("register", "svc", svc, "key", key)
	return nil
}

// SetTTL sets key-value with ttl time
func SetTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	if cliKv == nil {
		return ErrClientKvNil
	}
	if ttl > 0 {
		_, err := cliKv.Set(ctx, key, value, &client.SetOptions{
			TTL:       ttl,
			PrevExist: client.PrevIgnore,
		})
		return err
	}
	_, err := cliKv.Create(ctx, key, value)
	return err
}

func registerService(svc Service, ttl time.Duration) (string, string, error) {
	key := fmt.Sprintf("%s/%s/%s", RegisterPrefix, svc.Name, svc.Endpoint)
	value, err := json.Marshal(svc)
	if err != nil {
		return "", "", err
	}
	return key, string(value), SetTTL(context.Background(), key, string(value), ttl)
}

func setConfig(svc Service, config interface{}) error {
	key := svc.KeyPrefix() + "config"
	value, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return Set(context.Background(), key, string(value))
}

func registerLog(svc Service, value string) error {
	if cliKv == nil {
		return ErrClientKvNil
	}
	key := fmt.Sprintf("%s/%s/%s", RegisterLogPrefix, svc.Name, svc.Endpoint)
	_, err := cliKv.Set(context.Background(), key, value, &client.SetOptions{
		PrevExist: client.PrevIgnore,
	})
	return err
}

var (
	isServiceStop int64
)

func keepHeart(key, value string, heart, ttl time.Duration) {
	ticker := time.NewTicker(heart)
	defer ticker.Stop()
	for {
		<-ticker.C
		if atomic.LoadInt64(&isServiceStop) > 0 {
			log.Info("heartbeat-stop", "key", key)
			return
		}
		if err := SetTTL(context.Background(), key, value, ttl); err != nil {
			log.Warn("heartbeat-error", "key", key, "error", err)
		}
		// r.logger.Debug("heartbeat-ok")
	}
}

// Deregister deregisters current service
func Deregister() {
	if cliKv == nil {
		return
	}
	if serviceKey == "" {
		return
	}
	atomic.StoreInt64(&isServiceStop, 1)
	if err := Del(context.Background(), serviceKey); err != nil {
		log.Warn("deregister-error", "error", err, "key", serviceKey)
		return
	}
	log.Info("deregister", "key", serviceKey)
	serviceKey = ""
}
