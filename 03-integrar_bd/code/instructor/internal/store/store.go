package store

import (
	"fmt"

	"github.com/google/uuid"
)

type store interface {
	GetAllUsers() []User
}

type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"biography"`
	Id        string `json:"id"`
}

func convertMapToSlice[K comparable, V any](inputMap map[K]V) []V {
	var result []V
	for _, value := range inputMap {
		result = append(result, value)
	}
	return result
}

type UserRepo map[string]User

func (ur UserRepo) FindAll() []User {
	return convertMapToSlice(ur)
}

func (ur UserRepo) FindById(id string) (User, error) {
	user, ok := ur[id]
	if !ok {
		return User{}, fmt.Errorf("user with id: %s not found", id)
	}

	return user, nil
}

func (ur UserRepo) Insert(firstName, lastName, bio string) (User, error) {
	newId, err := uuid.NewRandom()
	if err != nil {
		return User{}, err
	}

	user := User{
		firstName,
		lastName,
		bio,
		newId.String(),
	}

	ur[newId.String()] = user

	return user, nil
}

func (ur UserRepo) Update(id string, u User) (User, error) {
	user, ok := ur[id]
	if !ok {
		return User{}, fmt.Errorf("user with id: %s not found", id)
	}

	if u.FirstName != "" {
		user.FirstName = u.FirstName
	}
	if u.LastName != "" {
		user.LastName = u.LastName
	}
	if u.Bio != "" {
		user.Bio = u.Bio
	}

	ur[id] = user

	return user, nil
}

func (ur UserRepo) Delete(id string) (User, error) {
	user, ok := ur[id]
	if !ok {
		return User{}, fmt.Errorf("user with id: %s not found", id)
	}

	delete(ur, id)

	return user, nil
}
