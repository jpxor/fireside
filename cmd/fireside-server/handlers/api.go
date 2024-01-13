package handlers

import (
	"fireside/app/db"
	"strings"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

var name = "World"

type userFormData struct {
	Name  string `form:"name"`
	Email string `form:"email"`
	Passw string `form:"password"`
}

func UserCreate(c *fiber.Ctx) error {
	var data userFormData
	err := c.BodyParser(&data)
	if err != nil || len(data.Passw) == 0 {
		return c.Status(fiber.StatusOK).SendString("bad request: failed to parse form data")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.Passw), bcrypt.DefaultCost)
	if err != nil {
		log.Println("UserCreate:bcrypt.GenerateFromPassword", err)
		return c.Status(fiber.StatusOK).SendString("server error: try again")
	}

	emailCopy := CopyString(data.Email)
	uid, err := db.SaveUnverifiedUser(emailCopy, hash)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("that email address is already in use - please login or reset password")
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "uid-unverified"
	cookie.Value = uid
	cookie.Expires = time.Now().Add(10 * time.Minute)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	// FUTURE: send email with link to confirmation page (verifies email)
	c.Set("HX-Redirect", "/confirm-password")
	return c.SendStatus(fiber.StatusOK)
}

func UserVerify(c *fiber.Ctx) error {
	var data userFormData
	err := c.BodyParser(&data)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("bad request: failed to parse form data")
	}

	uid := c.Cookies("uid-unverified")
	user, ok := db.GetUnverifiedUser(uid)

	if !ok {
		return c.Status(fiber.StatusOK).SendString("verification timedout - please try creating a new account")
	}
	if !user.CheckPassword([]byte(data.Passw)) {
		return c.Status(fiber.StatusOK).SendString("passwords don't match")
	}

	user.Name = CopyString(data.Name)
	err = db.SaveUser(user)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("server error: try again")
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "uid"
	cookie.Value = uid
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	c.Set("HX-Redirect", "/dashboard")
	return c.SendStatus(fiber.StatusOK)
}

func UserLogin(c *fiber.Ctx) error {
	var data userFormData
	err := c.BodyParser(&data)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("bad request: failed to parse form data")
	}

	user, _ := db.GetUser(data.Email)
	if !user.CheckPassword([]byte(data.Passw)) {
		return c.Status(fiber.StatusOK).SendString("bad email or password")
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "uid"
	cookie.Value = user.ID
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	c.Set("HX-Redirect", "/dashboard")
	return c.SendStatus(fiber.StatusOK)
}

func ResetLoginExpirationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if cookieValue := c.Cookies("login"); cookieValue != "" {
			newCookie := new(fiber.Cookie)
			newCookie.Name = "login"
			newCookie.Value = cookieValue
			newCookie.Expires = time.Now().Add(24 * time.Hour)
			c.Cookie(newCookie)
		}
		return c.Next()
	}
}

func DebugListUsers(c *fiber.Ctx) error {
	if c.Cookies("uid") != "" {
		return c.SendString(db.DebugListUsers())
	}
	return c.SendStatus(fiber.StatusForbidden)
}

func CopyString(str string) string {
	sb := strings.Builder{}
	sb.WriteString(str)
	return sb.String()
}
