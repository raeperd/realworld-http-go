package realworld

import "context"

// WARN: Need password hashing in production
type User struct {
	Name     string
	Email    string
	Password string
}

// TODO: Validate user
func (u User) Validate() error {
	return nil
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return UserService{repo: repo}
}

func (u UserService) CreateUser(ctx context.Context, user User) (User, error) {
	if err := user.Validate(); err != nil {
		return User{}, err
	}
	return u.repo.CreateUser(ctx, user)
}
