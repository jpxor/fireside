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
package main

import (
	"fmt"

	"github.com/jpxor/fireside/internal/app/fireside/auth"
	"github.com/jpxor/fireside/internal/app/fireside/server"
	"github.com/jpxor/fireside/internal/app/fireside/user"
	"github.com/skratchdot/open-golang/open"
)

type App struct {
	Users   user.Service
	Backend server.Service
	Auth    auth.Service
}

var Fireside App

func main() {

	fmt.Println("reading configs...")
	// TODO

	fmt.Println("initializing...")
	Fireside.Users = user.NewService("./users.json")
	Fireside.Auth = auth.New(Fireside.Users)
	Fireside.Backend = server.NewService(Fireside.Users, Fireside.Auth)

	fmt.Println("starting local private app server...")
	sigexit, port := Fireside.Backend.Start()

	fmt.Println("launching browser UI...")
	open.Run(fmt.Sprintf("http://localhost:%d/welcome", port))

	fmt.Println("waiting for app server to shutdown...")
	if sigexit != nil {
		<-sigexit
		close(sigexit)
	}
	fmt.Println("Bye!")
}
