package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/champbronc2/buzz-me/lightning"
	"github.com/champbronc2/buzz-me/model"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) CreatePost(c echo.Context) (err error) {
	u := &model.User{}
	p := &model.Post{
		ID:   bson.NewObjectId(),
		Paid: false,
	}
	if err = c.Bind(p); err != nil {
		return
	}

	// Validation
	if p.To == "" || p.Message == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid to or message fields"}
	}

	log.Println(p.To)

	// Find user from database
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("users").Find(bson.D{{"username", p.To}}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return
	}

	invoice, err := lightning.CreateInvoice(u.FeeRate)
	if err != nil {
		return err
	}

	p.Invoice = invoice

	// Save post in database
	if err = db.DB("buzzme").C("posts").Insert(p); err != nil {
		return
	}

	log.Println(invoice)

	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) FetchPost(c echo.Context) (err error) {
	userID := userIDFromToken(c)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
	}

	// Retrieve posts from database
	posts := []*model.Post{}
	db := h.DB.Clone()
	if err = db.DB("buzzme").C("posts").
		Find(bson.M{"to": userID}).
		Skip((page - 1) * limit).
		Limit(limit).
		All(&posts); err != nil {
		return
	}
	defer db.Close()

	return c.JSON(http.StatusOK, posts)
}
