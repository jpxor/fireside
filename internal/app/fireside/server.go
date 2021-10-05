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
	"time"

	"github.com/gofiber/fiber/v2"
)

type ServerContext struct {
	App            *fiber.App
	KillTime       time.Time
	WatchdogActive bool
}

var serverContext = ServerContext{}

func KeepAlive() {
	serverContext.KillTime = time.Now().Add(10 * time.Second)
}

func killerWatchdog() {
	for {
		time.Sleep(10 * time.Second)
		if serverContext.KillTime.Before(time.Now()) && serverContext.App != nil {
			serverContext.App.Shutdown()
			break
		}
	}
}

func TryPrivateServer(port int) (bool, chan (interface{})) {

	if !serverContext.WatchdogActive {
		serverContext.KillTime = time.Now().Add(10 * time.Second)
		serverContext.WatchdogActive = true
		go killerWatchdog()
	}

	sigfail := make(chan interface{})
	sigexit := make(chan interface{})

	go RunPrivateServer(port, sigexit, sigfail)
	var success bool

	select {
	// if the server successfully starts, it will
	// not send pulse on the sigfail channel, so
	// timeout means success
	case <-sigfail:
		close(sigexit)
		success = false
		sigexit = nil
	case <-time.After(time.Second):
		success = true
	}
	close(sigfail)
	return success, sigexit
}

func RunPrivateServer(port int, sigexit, sigfail chan (interface{})) {

	app := fiber.New()
	serverContext.App = app

	app.Static("/", "./web/assets")
	app.Static("/", "./web/static")
	app.Static("/", "./web/app")

	// app.Get("/welcome", func(c *fiber.Ctx) error {
	// 	return c.SendString("Hello, World!")
	// })

	start := time.Now()
	err := app.Listen(fmt.Sprintf(":%d", port))

	if time.Since(start) < time.Duration(500*time.Millisecond) {
		// if the app fails to listen, it will
		// return right away with error
		log.Println(err)
		sigfail <- 0
		return
	}
	if err != nil {
		log.Println(err)
	}
	// wakes main thread to exit app
	sigexit <- 0
}
