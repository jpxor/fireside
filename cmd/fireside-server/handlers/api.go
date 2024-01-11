package handlers

import "github.com/gofiber/fiber/v2"

var name = "World"

func GetName(c *fiber.Ctx) error {
	return c.SendString(name)
}
