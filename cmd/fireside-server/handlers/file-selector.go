package handlers

import (
	"fireside/app"
	"fmt"
	"io/fs"
	"path"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
)

type pathCrumbs struct {
	Name string
	Path string
}

type fileSelectorRenderData struct {
	Path           string
	SelectedFile   string
	PathCrumbs     []pathCrumbs
	DirEnts        []fs.DirEntry
	ReloadRecentTx bool
	Error          error
}

func RenderFileSelector(c *fiber.Ctx) error {
	sess, err := parseSessionCookie(c.Cookies("session"))
	if err != nil {
		c.ClearCookie("session")
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusOK)
	}
	return RenderFileSelectorWithSession(c, sess)
}

func RenderFileSelectorWithSession(c *fiber.Ctx, sess sessCookieData) error {
	makeCrumbs := func(dirpath string) []pathCrumbs {
		splits := strings.Split(dirpath, "/")
		crumbs := make([]pathCrumbs, 0, len(splits)+1)
		crumbs = append(crumbs, pathCrumbs{
			Name: "root",
			Path: "",
		})
		for i, name := range splits {
			if name == "." {
				continue
			}
			crumbs = append(crumbs, pathCrumbs{
				Name: name,
				Path: path.Join(crumbs[i].Path, name),
			})
		}
		return crumbs
	}
	dirPath := path.Clean(c.Params("*"))
	data := fileSelectorRenderData{
		Path:           dirPath,
		PathCrumbs:     makeCrumbs(dirPath),
		SelectedFile:   sess.SelectedFile,
		ReloadRecentTx: c.Locals("ReloadRecentTx") == true,
	}
	dirents, err := app.DirectoryListing(sess.ID, dirPath)
	if err != nil {
		data.Error = err
		return c.Render("file-selector.html", data)
	}
	data.DirEnts = dirents
	return c.Render("file-selector.html", data)
}

type fileSelectorNewData struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func FileSelectorNew(c *fiber.Ctx) error {
	sess, err := parseSessionCookie(c.Cookies("session"))
	if err != nil {
		c.ClearCookie("session")
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusOK)
	}
	dirPath := path.Clean(c.Params("*"))

	var data fileSelectorNewData
	if err := c.BodyParser(&data); err != nil {
		data := fileSelectorRenderData{
			Path:  dirPath,
			Error: err,
		}
		return c.Render("file-selector.html", data)
	}

	switch data.Type {
	case "folder":
		err = app.CreateFolder(sess.ID, dirPath, data.Name)
	case "journal":
		err = app.CreateJournal(sess.ID, dirPath, data.Name)
	default:
		err = fmt.Errorf("type not implemented")
	}
	if err != nil {
		data := fileSelectorRenderData{
			Path:  dirPath,
			Error: err,
		}
		return c.Render("file-selector.html", data)
	}
	return RenderFileSelectorWithSession(c, sess)
}

type fileSelectorSelectData struct {
	Name string `json:"name"`
}

func FileSelectorSelect(c *fiber.Ctx) error {
	sess, err := parseSessionCookie(c.Cookies("session"))
	if err != nil {
		c.ClearCookie("session")
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusOK)
	}
	file := c.Params("*")

	var data fileSelectorSelectData
	if err := c.BodyParser(&data); err != nil {
		data := fileSelectorRenderData{
			Path:  file,
			Error: err,
		}
		return c.Render("file-selector.html", data)
	}
	sess.SelectedFile = data.Name
	err = updateSessionCookie(c, sess)
	if err != nil {
		data := fileSelectorRenderData{
			Path:  file,
			Error: err,
		}
		return c.Render("file-selector.html", data)
	}
	c.Locals("ReloadRecentTx", true)
	return RenderFileSelectorWithSession(c, sess)
}
