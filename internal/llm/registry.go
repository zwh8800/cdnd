package llm

import (
	"fmt"
	"sync"
)

// Registry 管理可用的 LLM 提供者。
type Registry struct {
	mu              sync.RWMutex
	providers       map[string]Provider
	defaultProvider string
}

// NewRegistry 创建一个新的提供者注册表。
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register 向注册表添加一个提供者。
func (r *Registry) Register(name string, provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider

	// 如果是第一个提供者，则设为默认
	if len(r.providers) == 1 {
		r.defaultProvider = name
	}

	return nil
}

// Unregister 从注册表移除一个提供者。
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.providers, name)

	// 如果需要，更新默认提供者
	if r.defaultProvider == name {
		r.defaultProvider = ""
		for name := range r.providers {
			r.defaultProvider = name
			break
		}
	}
}

// Get 根据名称获取提供者。
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// Default 返回默认提供者。
func (r *Registry) Default() (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.defaultProvider == "" {
		return nil, fmt.Errorf("no default provider set")
	}

	provider, exists := r.providers[r.defaultProvider]
	if !exists {
		return nil, fmt.Errorf("default provider %s not found", r.defaultProvider)
	}

	return provider, nil
}

// SetDefault 设置默认提供者。
func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	r.defaultProvider = name
	return nil
}

// List 返回所有已注册的提供者名称。
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// 全局注册表实例
var globalRegistry = NewRegistry()

// RegisterProvider 全局注册一个提供者。
func RegisterProvider(name string, provider Provider) error {
	return globalRegistry.Register(name, provider)
}

// GetProvider 从全局注册表获取一个提供者。
func GetProvider(name string) (Provider, error) {
	return globalRegistry.Get(name)
}

// DefaultProvider 返回全局注册表中的默认提供者。
func DefaultProvider() (Provider, error) {
	return globalRegistry.Default()
}

// SetDefaultProvider 设置全局注册表中的默认提供者。
func SetDefaultProvider(name string) error {
	return globalRegistry.SetDefault(name)
}

// ListProviders 返回全局注册表中所有已注册的提供者名称。
func ListProviders() []string {
	return globalRegistry.List()
}
