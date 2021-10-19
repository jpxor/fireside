/*
 *  Copyright © 2021 Josh Simonot
 *
 *  fireside is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  fireside is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with fireside. If not, see <https://www.gnu.org/licenses/>.
 */
package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jpxor/fireside/internal/app/fireside/user"
	"golang.org/x/crypto/bcrypt"
)

type authImpl struct {
	Users      user.Service
	privateKey []byte
}

func generatePrivateKey() []byte {
	const pKeySize = 32
	pkey := make([]byte, pKeySize)
	_, err := rand.Read(pkey)
	if err != nil {
		log.Fatalln("failed to create private key:", err)
	}
	return pkey
}

func New(users user.Service) Service {
	return &authImpl{
		Users:      users,
		privateKey: generatePrivateKey(),
	}
}

func (auth *authImpl) Authenticate(uid, password string) (string, error) {
	const authDuration = 10 * time.Millisecond
	start := time.Now()

	usr, usrerr := auth.Users.Get(uid)
	if usr == nil {
		usr = &user.Model{}
	}
	hasherr := bcrypt.CompareHashAndPassword([]byte(usr.Hash), []byte(password))

	elapsed := time.Since(start)
	if elapsed < authDuration {
		time.Sleep(authDuration - elapsed)
	}
	if hasherr != nil || usrerr != nil {
		return "", errors.New("wrong user or password")
	}

	claims := Claims{
		UserID: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Issuer:    "fireside local private app server",
			Subject:   usr.Name,
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(auth.privateKey)
	return token, err
}

func (auth *authImpl) Refresh(tokenstr string) (string, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenstr, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return auth.privateKey, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}
	if claims.StandardClaims.ExpiresAt < time.Now().Unix() {
		return "", errors.New("invalid expired")
	}
	newClaims := Claims{
		UserID: claims.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Issuer:    "fireside local private app server",
		},
	}
	newToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims).SignedString(auth.privateKey)
	return newToken, err
}

func (auth *authImpl) Authorize(token, resource string) (success bool) {
	return false
}

func (auth *authImpl) Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	return string(hash), err
}
