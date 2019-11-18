package handler

import (
	"net/http"
	"strconv"

	"github.com/champbronc2/buzz-me/model"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	Handler struct {
		DB *mgo.Session
	}
)

const (
	// Key (Should come from somewhere else).
	Key = "secret"
)

func (h *Handler) Index(c echo.Context) (err error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 4
	}

	// Retrieve users from database
	users := []*model.PublicUser{}
	db := h.DB.Clone()
	if err = db.DB("buzzme").C("users").
		Find(nil).
		Skip((page - 1) * limit).
		Limit(limit).
		All(&users); err != nil {
		return
	}
	defer db.Close()

	// Retrieve 10 recent posts
	posts := []*model.Post{}
	db = h.DB.Clone()
	if err = db.DB("buzzme").C("posts").
		Find(bson.D{{"public", "on"}, {"paid", true}}).
		Skip((page - 1) * 10).
		Sort("-$natural").
		Limit(10).
		All(&posts); err != nil {
		return
	}
	defer db.Close()

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"users": users,
		"posts": posts,
	})
}
