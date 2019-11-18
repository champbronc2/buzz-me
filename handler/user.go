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
	if u.FeeRate == 0 {
		// 1000 sats/message default
		u.FeeRate = 1000
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
		Find(bson.M{"username": u.Username, "password": u.Password}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid username or password"}
		}
		return
	}

	//-----
	// JWT
	//-----

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	expTime := time.Now().Add(time.Hour * 72)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = u.ID
	claims["username"] = u.Username
	claims["exp"] = expTime.Unix()

	// Generate encoded token and send it as response
	u.Token, err = token.SignedString([]byte(Key))
	if err != nil {
		return err
	}

	u.Password = "" // Don't send password
	u.ID = ""       // Don't send ID

	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:    "Authorization",
		Value:   u.Token,
		Expires: expTime,
	})

	return c.JSON(http.StatusOK, u)
}

func (h *Handler) ViewLogin(c echo.Context) (err error) {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{})
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
		"Avatar":   users[0].Avatar,
		"twitter":  users[0].Twitter,
		"pgp_key":  users[0].PgpKey,
	})
}

func (h *Handler) Dashboard(c echo.Context) (err error) {
	username := usernameFromToken(c)
	u := new(model.User)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10000
	}

	// Find user
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("users").
		Find(bson.M{"username": username}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid token"}
		}
		return
	}

	// Retrieve posts from database
	posts := []*model.Post{}
	db = h.DB.Clone()
	if err = db.DB("buzzme").C("posts").
		Find(bson.D{{"to", username}, {"paid", true}}).
		Skip((page - 1) * limit).
		Sort("-$natural").
		Limit(limit).
		All(&posts); err != nil {
		return
	}
	defer db.Close()

	// Retrieve withdrawals from database
	withdrawals := []*model.Withdrawal{}
	db = h.DB.Clone()
	if err = db.DB("buzzme").C("withdrawals").
		Find(bson.D{{"username", username}}).
		Skip((page - 1) * limit).
		Sort("-$natural").
		Limit(limit).
		All(&withdrawals); err != nil {
		return
	}
	defer db.Close()

	// Calculate available balance
	balance := 0
	for _, post := range posts {
		if post.Paid && post.Read {
			balance = balance + post.Sats
		}
	}

	withdrew := 0
	for _, withdrawal := range withdrawals {
		withdrew = withdrew + withdrawal.Sats
	}

	balance = balance - withdrew

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"user":        u,
		"posts":       posts,
		"balance":     balance,
		"withdrawals": withdrawals,
	})
}

func (h *Handler) UpdateUser(c echo.Context) (err error) {
	// Bind
	u := &model.User{ID: bson.NewObjectId()}
	if err = c.Bind(u); err != nil {
		return
	}

	// Validate
	if u.Username == "" || u.Password == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid username, email or password"}
	}
	if u.FeeRate == 0 {
		// 1000 sats/message default
		u.FeeRate = 1000
	}

	// Save user
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("users").Insert(u); err != nil {
		return
	}

	return c.JSON(http.StatusCreated, u)
}

func (h *Handler) CreateWithdrawal(c echo.Context) (err error) {
	username := usernameFromToken(c)
	u := new(model.User)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10000
	}

	// Bind
	w := &model.Withdrawal{
		Username: username,
		ID:       bson.NewObjectId(),
		Sats:     1000,
		Paid:     false,
	}
	if err = c.Bind(w); err != nil {
		return
	}

	// Validate Lightning Invoice and extract sats

	// Retrieve posts from database
	posts := []*model.Post{}
	db := h.DB.Clone()
	if err = db.DB("buzzme").C("posts").
		Find(bson.D{{"to", username}, {"paid", true}}).
		Skip((page - 1) * limit).
		Sort("-$natural").
		Limit(limit).
		All(&posts); err != nil {
		return
	}
	defer db.Close()

	// Retrieve withdrawals from database
	withdrawals := []*model.Withdrawal{}
	db = h.DB.Clone()
	if err = db.DB("buzzme").C("withdrawals").
		Find(bson.D{{"username", username}}).
		Skip((page - 1) * limit).
		Sort("-$natural").
		Limit(limit).
		All(&withdrawals); err != nil {
		return
	}
	defer db.Close()

	// Calculate available balance
	balance := 0
	for _, post := range posts {
		if post.Paid && post.Read {
			balance = balance + post.Sats
		}
	}

	withdrew := 0
	for _, withdrawal := range withdrawals {
		withdrew = withdrew + withdrawal.Sats
	}

	availableBalance := balance - withdrew

	if w.Sats > availableBalance {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "you don't have that many sats"}
	}

	// Save withdrawal request
	db = h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("withdrawals").Insert(w); err != nil {
		return
	}

	// Try to send payment?? Maybe manual at first

	// Refresh withdrawals
	withdrawals = []*model.Withdrawal{}
	db = h.DB.Clone()
	if err = db.DB("buzzme").C("withdrawals").
		Find(bson.D{{"username", username}}).
		Skip((page - 1) * limit).
		Sort("-$natural").
		Limit(limit).
		All(&withdrawals); err != nil {
		return
	}
	defer db.Close()

	// Find user
	db = h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("users").
		Find(bson.M{"username": username}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid token"}
		}
		return
	}

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"user":        u,
		"posts":       posts,
		"balance":     balance,
		"withdrawals": withdrawals,
	})
}

func usernameFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["username"].(string)
}
