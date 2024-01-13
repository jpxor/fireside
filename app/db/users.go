package db

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email string
	Name  string
	ID    string
	hash  []byte
}

type UserMap struct {
	kv map[string]User
	sync.Mutex
}

var (
	users = UserMap{
		kv: make(map[string]User),
	}
	unverifiedUsers = UserMap{
		kv: make(map[string]User),
	}
	userMutex sync.Mutex
)

func (u User) CheckPassword(passw []byte) bool {
	return bcrypt.CompareHashAndPassword(u.hash, passw) == nil
}

func GetUser(email string) (user User, ok bool) {
	userMutex.Lock()
	defer userMutex.Unlock()
	user, ok = users.kv[email]
	return
}

func GetUnverifiedUser(uid string) (user User, ok bool) {
	userMutex.Lock()
	defer userMutex.Unlock()
	user, ok = unverifiedUsers.kv[uid]
	return
}

func UserEmailExists(email string) bool {
	userMutex.Lock()
	defer userMutex.Unlock()
	_, exists := users.kv[email]
	return exists
}

func SaveUser(user User) error {
	userMutex.Lock()
	defer userMutex.Unlock()
	users.kv[user.Email] = user
	delete(unverifiedUsers.kv, user.ID)
	return nil
}

func SaveUnverifiedUser(email string, hash []byte) (string, error) {
	userMutex.Lock()
	defer userMutex.Unlock()

	_, exists := users.kv[email]
	if exists {
		return "", fmt.Errorf("email already in use")
	}
	user := User{
		Email: email,
		ID:    xid.New().String(),
		hash:  hash,
	}
	unverifiedUsers.kv[user.ID] = user

	// launch a go routine that will clear the unvalidated user
	// after ~10 minutes (for low volume)
	go func() {
		time.Sleep(10 * time.Minute)
		userMutex.Lock()
		defer userMutex.Unlock()
		delete(unverifiedUsers.kv, user.ID)
	}()

	return user.ID, nil
}

func DebugListUsers() string {
	userMutex.Lock()
	defer userMutex.Unlock()
	jstr0, _ := json.MarshalIndent(users.kv, "", "    ")
	jstr1, _ := json.MarshalIndent(unverifiedUsers.kv, "", "    ")
	return fmt.Sprintf("verified\r\n%s\r\n\nunverified\r\n%s\r\n", jstr0, jstr1)

}
