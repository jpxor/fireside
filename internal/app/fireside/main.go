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
package fireside

import (
	"fmt"
	"log"

	"github.com/skratchdot/open-golang/open"
)

func Run() {

	fmt.Println("reading configs...")
	// TODO

	fmt.Println("starting local private app server...")
	var port int
	var success bool
	var sigexit chan (interface{})

	for _, tryport := range []int{80, 8080, 8081, 8082, 8083, 8084} {
		success, sigexit = TryPrivateServer(tryport)
		if success {
			fmt.Printf("server will shutdown when connection with client is lost\n")
			port = tryport
			break
		}
	}

	if !success {
		fmt.Println("failed to start server")
		log.Fatal("ABORT")
	}

	fmt.Println("launching browser UI...")
	open.Run(fmt.Sprintf("http://localhost:%d/welcome", port))

	// waiting for server to lose connection with
	// browser before exiting app
	<-sigexit
	close(sigexit)
	fmt.Println("Bye!")
}
