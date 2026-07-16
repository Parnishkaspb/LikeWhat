package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Parnishkaspb/LikeWhat/internal/likewhat/tobacco"
)

// Memory is a concurrency-safe in-memory TobaccoRepository implementation.
// It is useful for local development and tests; data is lost when the process stops.
type Memory struct {
	mu     sync.RWMutex
	items  map[string]tobacco.Tobacco
	order  []string
	nextID uint64
	now    func() time.Time
}

func NewMemory() *Memory {
	return &Memory{
		items: make(map[string]tobacco.Tobacco),
		now:   time.Now,
	}
}

func (r *Memory) Create(ctx context.Context, item tobacco.Tobacco) (tobacco.Tobacco, error) {
	if err := ctx.Err(); err != nil {
		return tobacco.Tobacco{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	now := r.now().UTC()
	item.ID = fmt.Sprintf("tobacco-%d", r.nextID)
	item.CreatedAt = now
	item.UpdatedAt = now
	r.items[item.ID] = item
	r.order = append(r.order, item.ID)
	return item, nil
}

func (r *Memory) Get(ctx context.Context, id string) (tobacco.Tobacco, error) {
	if err := ctx.Err(); err != nil {
		return tobacco.Tobacco{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok || item.DeletedAt != nil {
		return tobacco.Tobacco{}, ErrNotFound
	}
	return item, nil
}

func (r *Memory) List(ctx context.Context, filter tobacco.ListFilter) ([]tobacco.Tobacco, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	manufactureIDs := make(map[string]struct{}, len(filter.ManufactureIDs))
	for _, id := range filter.ManufactureIDs {
		manufactureIDs[id] = struct{}{}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]tobacco.Tobacco, 0, len(r.order))
	for _, id := range r.order {
		item := r.items[id]
		if item.DeletedAt != nil || !matches(item, filter.Taste, manufactureIDs) {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func matches(item tobacco.Tobacco, taste string, manufactureIDs map[string]struct{}) bool {
	if taste != "" && !strings.EqualFold(item.Taste, taste) {
		return false
	}
	if len(manufactureIDs) == 0 {
		return true
	}
	_, ok := manufactureIDs[item.Manufacture.ID]
	return ok
}
