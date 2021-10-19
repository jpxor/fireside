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

import "github.com/golang-jwt/jwt"

type Claims struct {
	UserID string
	jwt.StandardClaims
}

type Service interface {
	Authenticate(uid, password string) (token string, err error)
	Authorize(token, resource string) (success bool)
	Refresh(token string) (newToken string, err error)
	Hash(plain string) (hash string, err error)
}
