package store

import (
	"sync"

	"github.com/dante-gpu/dante-backend/provider-registry-service/internal/models"
	"github.com/google/uuid"
)

// InMemoryProviderStore is a simple in-memory store for providers.
// I need to make this thread-safe for concurrent access.
type InMemoryProviderStore struct {
	mu        sync.RWMutex
	providers map[uuid.UUID]*models.Provider
}

// NewInMemoryProviderStore creates a new in-memory provider store.
func NewInMemoryProviderStore() *InMemoryProviderStore {
	return &InMemoryProviderStore{
		providers: make(map[uuid.UUID]*models.Provider),
	}
}

// AddProvider adds a new provider to the store.
func (s *InMemoryProviderStore) AddProvider(provider *models.Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// I should check if a provider with this ID already exists, though UUIDs should be unique.
	if _, exists := s.providers[provider.ID]; exists {
		return models.ErrProviderAlreadyExists // I'll need to define this error type
	}
	s.providers[provider.ID] = provider
	return nil
}

// GetProvider retrieves a provider by its ID.
func (s *InMemoryProviderStore) GetProvider(id uuid.UUID) (*models.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[id]
	if !exists {
		return nil, models.ErrProviderNotFound // I'll need to define this error type
	}
	return provider, nil
}

// ListProviders returns a list of all providers, with optional filtering (TODO).
// For now, it returns all providers.
func (s *InMemoryProviderStore) ListProviders( /* filters map[string]string */ ) ([]*models.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*models.Provider, 0, len(s.providers))
	for _, provider := range s.providers {
		list = append(list, provider)
	}
	return list, nil
}

// UpdateProvider updates an existing provider in the store.
func (s *InMemoryProviderStore) UpdateProvider(id uuid.UUID, updatedProvider *models.Provider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.providers[id]
	if !exists {
		return models.ErrProviderNotFound
	}
	// I should ensure the ID is not changed during an update.
	updatedProvider.ID = id
	s.providers[id] = updatedProvider
	return nil
}

// DeleteProvider removes a provider from the store.
func (s *InMemoryProviderStore) DeleteProvider(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.providers[id]; !exists {
		return models.ErrProviderNotFound
	}
	delete(s.providers, id)
	return nil
}

// UpdateProviderStatus updates the status of a specific provider.
func (s *InMemoryProviderStore) UpdateProviderStatus(id uuid.UUID, status models.ProviderStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider, exists := s.providers[id]
	if !exists {
		return models.ErrProviderNotFound
	}
	provider.UpdateStatus(status)
	s.providers[id] = provider // Re-assign to map if provider is a copy (though it's a pointer here)
	return nil
}

// UpdateProviderHeartbeat updates the LastSeenAt timestamp for a provider.
func (s *InMemoryProviderStore) UpdateProviderHeartbeat(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider, exists := s.providers[id]
	if !exists {
		return models.ErrProviderNotFound
	}
	provider.Heartbeat()
	s.providers[id] = provider
	return nil
}
