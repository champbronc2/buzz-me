package handler

import (
	"net/http"
	"strconv"

	"github.com/champbronc2/buzz-me/model"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
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

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"username0": users[0].Username,
		"username1": users[1].Username,
		"username2": users[2].Username,
		"username3": users[3].Username,

		"fee0": users[0].FeeRate,
		"fee1": users[1].FeeRate,
		"fee2": users[2].FeeRate,
		"fee3": users[3].FeeRate,
	})
}
