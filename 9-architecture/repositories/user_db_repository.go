package repositories

import (
	"stepikGoWebServices/handlers/entities"
	"strconv"
	"sync"
	"time"
)

type UserDbRepository struct {
	users     map[string]*entities.User
	usersById map[string]*entities.User

	userMutex *sync.RWMutex
}

func NewUserDbRepository() *UserDbRepository {
	return &UserDbRepository{
		users:     make(map[string]*entities.User),
		usersById: make(map[string]*entities.User),
		userMutex: &sync.RWMutex{},
	}
}

func (db *UserDbRepository) AddUser(
	email string,
	username string,
	password string,
) *entities.User {
	var user = &entities.User{
		Id:        db.generateId(),
		Email:     email,
		Username:  username,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db.users[user.Email] = user
	db.usersById[user.Id] = user

	return user
}

func (db *UserDbRepository) GetUserByEmail(email string) *entities.User {
	db.userMutex.RLock()
	defer db.userMutex.RUnlock()
	return db.users[email]
}

func (db *UserDbRepository) GetUserById(id string) *entities.User {
	db.userMutex.RLock()
	defer db.userMutex.RUnlock()
	return db.usersById[id]
}

func (db *UserDbRepository) IsUserExists(email string) bool {
	db.userMutex.RLock()
	defer db.userMutex.RUnlock()

	var _, exists = db.users[email]
	return exists
}

func (db *UserDbRepository) UpdateUser(
	oldUser *entities.User,
	newUser *entities.User,
) {
	db.userMutex.Lock()
	defer db.userMutex.Unlock()

	delete(db.users, oldUser.Email)
	delete(db.usersById, oldUser.Id)

	db.users[newUser.Email] = newUser
	db.usersById[newUser.Id] = newUser
}

func (db *UserDbRepository) generateId() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}
