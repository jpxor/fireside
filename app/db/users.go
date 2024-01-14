package db

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rs/xid"
	bolt "go.etcd.io/bbolt"
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

var userdb *bolt.DB
var unverifiedUsers = UserMap{
	kv: make(map[string]User),
}

func InitUserDB(path string) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			log.Fatal(err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("hashes"))
		if err != nil {
			log.Fatal(err)
		}
		return
	})
	userdb = db
}

func GetUser(email string) (user User, ok bool) {
	userdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		v := b.Get([]byte(email))
		if len(v) == 0 {
			return nil
		}
		err := json.Unmarshal(v, &user)
		if err != nil {
			log.Printf("GetUser:Unmarshal(%s): %s\r\n", string(v), err)
			return err
		}
		ok = true
		return nil
	})
	return
}

func CheckPassword(user User, passw string) bool {
	// when verifying user password, we have an unverified
	// user with hash (user not in db yet), but during
	// login we need to get the hash from the db
	if len(user.hash) == 0 {
		userdb.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("hashes"))
			user.hash = b.Get([]byte(user.Email))
			return nil
		})
	}
	return bcrypt.CompareHashAndPassword(user.hash, []byte(passw)) == nil
}

func GetUnverifiedUser(uid string) (user User, ok bool) {
	unverifiedUsers.Lock()
	defer unverifiedUsers.Unlock()
	user, ok = unverifiedUsers.kv[uid]
	return
}

func UserEmailExists(email string) bool {
	_, ok := GetUser(email)
	return ok
}

func SaveUser(user User) error {
	err := userdb.Update(func(tx *bolt.Tx) error {
		v, err := json.Marshal(user)
		if err != nil {
			return err
		}
		b := tx.Bucket([]byte("users"))
		err = b.Put([]byte(user.Email), v)
		if err != nil {
			return err
		}
		b = tx.Bucket([]byte("hashes"))
		err = b.Put([]byte(user.Email), user.hash)
		if err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		unverifiedUsers.Lock()
		defer unverifiedUsers.Unlock()
		delete(unverifiedUsers.kv, user.ID)
	}
	return err
}

func SaveUnverifiedUser(email string, hash []byte) (string, error) {
	unverifiedUsers.Lock()
	defer unverifiedUsers.Unlock()

	if UserEmailExists(email) {
		return "", fmt.Errorf("email already in use")
	}
	user := User{
		Email: email,
		ID:    xid.New().String(),
		hash:  hash,
	}
	unverifiedUsers.kv[user.ID] = user

	// launch a go routine that will clear the unverified user
	// after ~10 minutes (for low volume)
	go func() {
		time.Sleep(10 * time.Minute)
		unverifiedUsers.Lock()
		defer unverifiedUsers.Unlock()
		delete(unverifiedUsers.kv, user.ID)
	}()

	return user.ID, nil
}

func DebugListUsers() string {
	unverifiedUsers.Lock()
	defer unverifiedUsers.Unlock()

	users := make(map[string]User)
	userdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("users"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var u User
			err := json.Unmarshal(v, &u)
			if err != nil {
				log.Printf("DebugListUsers:Unmarshal(%s): %s\r\n", string(v), err)
				continue
			}
			users[u.Email] = u
		}
		return nil
	})

	jstr0, _ := json.MarshalIndent(users, "", "    ")
	jstr1, _ := json.MarshalIndent(unverifiedUsers.kv, "", "    ")
	return fmt.Sprintf("verified\r\n%s\r\n\nunverified\r\n%s\r\n", jstr0, jstr1)

}
