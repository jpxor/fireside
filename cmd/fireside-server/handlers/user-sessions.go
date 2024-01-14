package handlers

import (
	"encoding/json"
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

type sessCookieData struct {
	SelectedFile string
	Email        string
	ID           string
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

	emailCopy := copyString(data.Email)
	uid, err := db.SaveUnverifiedUser(emailCopy, hash)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("that email address is already in use - please login or reset password")
	}

	c.Cookie(newUnverifiedCookie(uid))

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

	uid := c.Cookies("session-unverified")
	user, ok := db.GetUnverifiedUser(uid)

	if !ok {
		return c.Status(fiber.StatusOK).SendString("verification timed out - please try creating a new account")
	}
	if !db.CheckPassword(user, data.Passw) {
		return c.Status(fiber.StatusOK).SendString("passwords don't match")
	}

	user.Name = copyString(data.Name)
	err = db.SaveUser(user)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("server error: try again")
	}

	err = newSessionCookie(c, user.Email, uid)
	if err != nil {
		// error creating session, but user is successfully verified,
		// redirect them to login
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusOK)
	}

	c.Set("HX-Redirect", "/dashboard")
	return c.SendStatus(fiber.StatusOK)
}

func UserLogin(c *fiber.Ctx) error {
	var data userFormData
	err := c.BodyParser(&data)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("bad request: failed to parse form data")
	}

	if !db.CheckPassword(db.User{Email: data.Email}, data.Passw) {
		return c.Status(fiber.StatusOK).SendString("bad email or password")
	}
	user, _ := db.GetUser(data.Email)

	err = newSessionCookie(c, user.Email, user.ID)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("server error: try again")
	}

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

func newUnverifiedCookie(uid string) *fiber.Cookie {
	cookie := new(fiber.Cookie)
	cookie.Name = "session-unverified"
	cookie.Value = uid
	cookie.Expires = time.Now().Add(10 * time.Minute)
	cookie.HTTPOnly = true
	return cookie
}

func newSessionCookie(c *fiber.Ctx, email, id string) error {
	buf, err := json.Marshal(sessCookieData{
		Email: email,
		ID:    id,
	})
	if err != nil {
		log.Println("newSessionCookie:", err)
		return err
	}
	cookie := new(fiber.Cookie)
	cookie.Name = "session"
	cookie.Value = string(buf)
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)
	return nil
}

func updateSessionCookie(c *fiber.Ctx, sess sessCookieData) error {
	buf, err := json.Marshal(sess)
	if err != nil {
		log.Println("updateSessionCookie:", err)
		return err
	}
	cookie := new(fiber.Cookie)
	cookie.Name = "session"
	cookie.Value = string(buf)
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)
	return nil
}

func parseSessionCookie(val string) (sess sessCookieData, err error) {
	err = json.Unmarshal([]byte(val), &sess)
	if err != nil {
		log.Printf("parseSessionCookie(%s): %s\r\n", val, err)
	}
	return
}

func copyString(str string) string {
	sb := strings.Builder{}
	sb.WriteString(str)
	return sb.String()
}
