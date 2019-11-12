package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/champbronc2/buzz-me/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) Signup(c echo.Context) (err error) {
	// Bind
	u := &model.User{ID: bson.NewObjectId()}
	if err = c.Bind(u); err != nil {
		return
	}

	// Validate
	if u.Username == "" || u.Password == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid username, email or password"}
	}
	if u.FeeRate == "" {
		// 1000 sats/message default
		u.FeeRate = "1000"
	}

	// Save user
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("users").Insert(u); err != nil {
		return
	}

	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) ViewSignup(c echo.Context) (err error) {
	return c.Render(http.StatusOK, "signup.html", map[string]interface{}{})
}

func (h *Handler) Login(c echo.Context) (err error) {
	// Bind
	u := new(model.User)
	if err = c.Bind(u); err != nil {
		return
	}

	// Find user
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("users").
		Find(bson.M{"email": u.Email, "password": u.Password}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid email or password"}
		}
		return
	}

	//-----
	// JWT
	//-----

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response
	u.Token, err = token.SignedString([]byte(Key))
	if err != nil {
		return err
	}

	u.Password = "" // Don't send password
	return c.JSON(http.StatusOK, u)
}

func (h *Handler) ListUsers(c echo.Context) (err error) {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
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

	return c.JSON(http.StatusOK, users)
}

func (h *Handler) FetchUser(c echo.Context) (err error) {
	username := c.Param("username")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
	}

	// Retrieve user from database
	users := []*model.PublicUser{}
	db := h.DB.Clone()
	if err = db.DB("buzzme").C("users").
		Find(bson.M{"username": username}).
		Skip((page - 1) * limit).
		Limit(limit).
		All(&users); err != nil {
		return
	}
	defer db.Close()

	return c.Render(http.StatusOK, "user.html", map[string]interface{}{
		"username": users[0].Username,
		"fee":      users[0].FeeRate,
		"avatar":   users[0].Avatar,
		"twitter":  users[0].Twitter,
		"pgp_key":  users[0].PgpKey,
	})
}

func userIDFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}
