package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/user"
	"context"
)

type UsersRepository struct {
	client *ent.Client
}

func NewUsersRepository(client *ent.Client) *UsersRepository {
	return &UsersRepository{client: client}
}

func (r *UsersRepository) GetUser(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Get(ctx, id)
}

func (r *UsersRepository) AddUser(ctx context.Context, name string, email string, password string) (*ent.User, error) {
	return r.client.User.Create().
		SetName(name).
		SetEmail(email).
		SetPassword(password).
		Save(ctx)
}

func (r *UsersRepository) GetUserByCredentials(ctx context.Context, email string, password string) (*ent.User, error) {
	foundUser, err := r.client.User.Query().
		Where(user.EmailEQ(email)).
		First(ctx)
	if err != nil {
		return nil, err
	}
	if foundUser.Password != password {
		return nil, &ent.NotFoundError{}
	}
	return foundUser, nil
}