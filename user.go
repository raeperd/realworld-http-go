package realworld

import "context"

// WARN: Need password hashing in production
type User struct {
	Profile
	Email    string
	Password string
}

type Profile struct {
	Username string
	Bio      string
	Image    string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByUsername(ctx context.Context, username string) (User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return UserService{repo: repo}
}

func (u UserService) CreateUser(ctx context.Context, user User) (User, error) {
	return u.repo.CreateUser(ctx, user)
}

func (u UserService) FindProfileByUsername(ctx context.Context, username string) (Profile, error) {
	user, err := u.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return Profile{}, err
	}
	return user.Profile, nil
}
