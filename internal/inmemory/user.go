package inmemory

import (
	"context"
	"sync"

	"github.com/raeperd/realworld"
)

// UserRepository implements [realworld.UserRepository]
type UserRepository struct {
	sync.RWMutex
	nextId uint
	memory map[uint]realworld.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		memory: make(map[uint]realworld.User),
	}
}

func (us *UserRepository) CreateUser(ctx context.Context, user realworld.User) (realworld.User, error) {
	us.Lock()
	us.memory[us.nextId] = user
	us.nextId += 1
	us.Unlock()
	return user, nil
}
