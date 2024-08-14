package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/raeperd/realworld"
)

// UserRepository implements [realworld.UserRepository]
type UserRepository struct {
	sync.RWMutex
	memory map[string]realworld.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		memory: make(map[string]realworld.User),
	}
}

func (us *UserRepository) CreateUser(ctx context.Context, user realworld.User) (realworld.User, error) {
	us.Lock()
	us.memory[user.Email] = user
	us.Unlock()
	return user, nil
}

func (us *UserRepository) FindUserByEmail(ctx context.Context, email string) (realworld.User, error) {
	us.RLock()
	defer us.RUnlock()
	user, ok := us.memory[email]
	if !ok {
		return realworld.User{}, fmt.Errorf("%w with email %s", realworld.ErrUserNotFound, email)
	}
	return user, nil
}
