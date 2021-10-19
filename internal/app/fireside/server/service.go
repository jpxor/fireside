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
package server

import (
	"fmt"
	"log"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/utils"
	"github.com/jpxor/fireside/internal/app/fireside/auth"
	"github.com/jpxor/fireside/internal/app/fireside/user"
)

type fiberHandler func(c *fiber.Ctx) error

type serverImpl struct {
	fiber  *fiber.App
	expire time.Time
	Users  user.Service
	Auth   auth.Service
}

func NewService(Users user.Service, Auth auth.Service) Service {
	serv := &serverImpl{
		expire: time.Now().Add(12 * time.Second),
		fiber:  fiber.New(),
		Users:  Users,
		Auth:   Auth,
	}
	serv.initHandlers()
	go serv.startWatchdog()
	return serv
}

func (s *serverImpl) keepAlive() {
	s.expire = time.Now().Add(10 * time.Second)
}

func (s *serverImpl) startWatchdog() {
	for {
		time.Sleep(10 * time.Second)
		if s.expire.Before(time.Now()) && s.fiber != nil {
			s.fiber.Shutdown()
			break
		}
	}
}

func (s *serverImpl) Start() (chan (interface{}), int) {
	timeout := time.Duration(250 * time.Millisecond)

	sigfail := make(chan interface{})
	sigexit := make(chan interface{})

	tryPorts := []int{80, 8080, 8081, 8082, 8083, 8084}
	for _, port := range tryPorts {

		go s.tryListen(timeout, port, sigexit, sigfail)
		select {
		case <-sigfail:
			continue
		case <-time.After(timeout):
			close(sigfail)
			return sigexit, port
		}
	}
	// failed to start server
	close(sigfail)
	close(sigexit)
	return nil, 0
}

func (s *serverImpl) tryListen(timeout time.Duration, port int, sigexit, sigfail chan (interface{})) {
	start := time.Now()

	// Listen is a blocking call and there is no way of notifying
	// the caller of success -- we can only send a signal on failure
	// so the caller should set a timeout and assume success if
	// it does not recieve the sigfail within that time
	err := s.fiber.Listen(fmt.Sprintf(":%d", port))

	if time.Since(start) < timeout {
		log.Println(err)
		sigfail <- 0
		return
	}
	if err != nil {
		log.Println(err)
	}
	// server was shutdown cleanly,
	// wake main thread to exit app
	sigexit <- 0
}

func (s *serverImpl) initHandlers() {

	s.fiber.Static("/", "./web/app")
	s.fiber.Static("/", "./web/assets")
	s.fiber.Static("/", "./web/static")

	s.fiber.Add(s.handleKeepAlive())

	api := s.fiber.Group("/api/v1")
	api.Add(s.handleGetUsers())
	api.Add(s.handlePostUser())
	api.Add(s.handleLogin())
}

func (s *serverImpl) handleKeepAlive() (string, string, fiberHandler) {
	return "PUT", "/api/keepalive",
		func(c *fiber.Ctx) error {
			s.keepAlive()
			return c.SendStatus(200)
		}
}

func (s *serverImpl) handlePostUser() (string, string, fiberHandler) {
	return "POST", "/users/:name",
		func(c *fiber.Ctx) error {
			hash, err := s.Auth.Hash("")
			if err != nil {
				return c.Status(500).SendString(err.Error())
			}
			err = s.Users.New(utils.ImmutableString(c.Params("name")), hash)
			if err != nil {
				return c.Status(409).SendString(err.Error())
			}
			return c.SendStatus(200)
		}
}

func (s *serverImpl) handleGetUsers() (string, string, fiberHandler) {
	return "GET", "/users",
		func(c *fiber.Ctx) error {
			type apiUser struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}
			response := []apiUser{}
			appendToResponse := func(usr *user.Model) error {
				response = append(response, apiUser{
					usr.Name,
					usr.ID,
				})
				return nil
			}
			s.Users.ForEach(appendToResponse)
			return c.JSON(response)
		}
}

func (s *serverImpl) handleLogin() (string, string, fiberHandler) {
	return "GET", "/auth/:id",
		func(c *fiber.Ctx) error {
			token, err := s.Auth.Authenticate(c.Params("id"), "")
			if err != nil {
				return c.Status(401).SendString("wrong user id or password")
			}
			type response struct {
				Token string `json:"token"`
			}
			return c.JSON(response{
				token,
			})
		}
}
