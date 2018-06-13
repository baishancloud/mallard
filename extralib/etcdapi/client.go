package etcdapi

import (
	"context"
	"errors"
	"time"

	"github.com/baishancloud/mallard/corelib/zaplog"
	"github.com/coreos/etcd/client"
)

var (
	log   = zaplog.Zap("etcd")
	cli   client.Client
	cliKv client.KeysAPI

	// ErrNoMachines means no machines to connect
	ErrNoMachines = errors.New("no-machines")
	// ErrClientNil means client is nil
	ErrClientNil = errors.New("client-is-nil")
	// ErrClientKvNil means client kv api is nil
	ErrClientKvNil = errors.New("kv-is-nil")
)

// MustSetClient sets clients, if error, panic
func MustSetClient(machines []string, user, password string, timeout time.Duration) {
	err := SetClient(machines, user, password, timeout)
	if err == nil || err == ErrNoMachines {
		return
	}
	log.Fatal("set-client-error", "error", err)
}

// SetClient sets etcd client
func SetClient(machines []string, user, password string, timeout time.Duration) error {
	if len(machines) == 0 {
		return ErrNoMachines
	}
	c, err := tryConnect(machines, user, password, timeout)
	if err != nil {
		return err
	}
	cli = c
	cliKv = client.NewKeysAPI(c)
	log.Info("set-client", "machines", machines)
	return nil
}

func tryConnect(machines []string, user, password string, timeout time.Duration) (client.Client, error) {
	transport := client.DefaultTransport
	return client.New(client.Config{
		Endpoints:               machines,
		Transport:               transport,
		HeaderTimeoutPerRequest: timeout,
		Username:                user,
		Password:                password,
	})
}

// Get gets value of key
func Get(ctx context.Context, key string) (string, error) {
	if cliKv == nil {
		return "", ErrClientKvNil
	}
	resp, err := cliKv.Get(ctx, key, &client.GetOptions{Recursive: false})
	if err != nil {
		return "", err
	}
	return resp.Node.Value, nil
}

// GetChildren gets children values of key
func GetChildren(ctx context.Context, key string) (map[string]string, error) {
	if cliKv == nil {
		return nil, ErrClientKvNil
	}
	resp, err := cliKv.Get(ctx, key, &client.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}
	if len(resp.Node.Nodes) == 0 {
		return nil, nil
	}
	values := make(map[string]string, len(resp.Node.Nodes))
	for _, node := range resp.Node.Nodes {
		values[node.Key] = node.Value
	}
	return values, nil
}

// Set sets key and value
func Set(ctx context.Context, key, value string) error {
	if cliKv == nil {
		return ErrClientKvNil
	}
	_, err := cliKv.Set(ctx, key, value, nil)
	return err
}

// Del removes key
func Del(ctx context.Context, key string) error {
	if cliKv == nil {
		return ErrClientKvNil
	}
	_, err := cliKv.Delete(ctx, key, &client.DeleteOptions{
		Recursive: true,
	})
	if err != nil {
		if client.IsKeyNotFound(err) {
			return nil
		}
	}
	return err
}

// Watch watches key and calls to channel
func Watch(ctx context.Context, key string, ch chan *client.Response) error {
	if cliKv == nil {
		return ErrClientKvNil
	}
	watch := cliKv.Watcher(key, &client.WatcherOptions{AfterIndex: 0, Recursive: true})
	for {
		resp, err := watch.Next(ctx)
		if err != nil {
			return err
		}
		if resp != nil && ch != nil {
			ch <- resp
		}
	}
}

// KeepWatch keeps watching key and retries to watch after error
func KeepWatch(ctx context.Context, key string, ch chan *client.Response, wait time.Duration) error {
	if cliKv == nil {
		return ErrClientKvNil
	}
	for {
		if err := Watch(ctx, key, ch); err != nil {
			log.Warn("keep-watch-break-error", "error", err, "key", key)
		}
		time.Sleep(wait)
	}
}
