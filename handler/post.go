package handler

import (
	"encoding/json"
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
		Read: true,
	}
	if err = c.Bind(p); err != nil {
		return
	}

	// Validation
	if p.To == "" || p.Message == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid to or message fields"}
	}

	// Find user from database
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("buzzme").C("users").Find(bson.D{{"username", p.To}}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return echo.ErrNotFound
		}
		return
	}

	// Generate invoice
	invoice, err := lightning.CreateInvoice(strconv.Itoa(u.FeeRate))
	if err != nil {
		return err
	}
	p.Invoice = invoice
	// p.Invoice = "{\"r_hash\":\"rGZUpNuwPGdysrfIEf7iso4PlrJHDUlIRCNsyDDO1E0=\",\"payment_request\":\"lntb1pwukphgpp543n9ffxmkq7xwu4jklyprlhzk28ql94jgux5jjzyydkvsvxw63xsdqqcqzpgraqptwft5jhckznertz77nu0zh4vd9afgwlyr352z4u8gty73mk88qruxk5nt6a33pn6l2prvgp9mq503kz5rjluh9afkzyw747vztcqd7t9qp\",\"add_index\":\"1\"}"
	p.Sats = u.FeeRate

	// Save post in database
	if err = db.DB("buzzme").C("posts").Insert(p); err != nil {
		return
	}

	return c.Redirect(http.StatusMovedPermanently, "/post/"+p.ID.Hex())
}

func (h *Handler) FetchPost(c echo.Context) (err error) {
	id := bson.ObjectIdHex(c.Param("id"))
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	// Defaults
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 100
	}

	// Retrieve post from database
	posts := []*model.Post{}
	db := h.DB.Clone()
	if err = db.DB("buzzme").C("posts").
		Find(bson.M{"_id": id}).
		Skip((page - 1) * limit).
		Limit(limit).
		All(&posts); err != nil {
		return
	}
	defer db.Close()

	p := posts[0]

	invoice := lightning.InvoiceResponse{}
	json.Unmarshal([]byte(p.Invoice), &invoice)

	return c.Render(http.StatusCreated, "post.html", map[string]interface{}{
		"id":             p.ID.Hex(),
		"from":           p.From,
		"to":             p.To,
		"paymentRequest": invoice.PaymentRequest,
		"message":        p.Message,
		"paid":           p.Paid,
	})
}

func (h *Handler) CheckPost(c echo.Context) (err error) {
	// Defaults
	var (
		page  = 1
		limit = 100
	)

	p := &model.Post{}
	if err = c.Bind(p); err != nil {
		return
	}

	// Validation
	if p.To == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid post"}
	}

	// Retrieve post from database
	posts := []*model.Post{}
	db := h.DB.Clone()
	if err = db.DB("buzzme").C("posts").
		Find(bson.M{"_id": bson.ObjectIdHex(p.To)}).
		Skip((page - 1) * limit).
		Limit(limit).
		All(&posts); err != nil {
		return
	}
	defer db.Close()

	p = posts[0]

	invoice := lightning.InvoiceResponse{}
	json.Unmarshal([]byte(p.Invoice), &invoice)

	invoicePaid, err := lightning.GetInvoicePaid(invoice)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "failed to get invoice"}
	}
	if invoicePaid {
		p.Paid = true
		log.Println("paid!")
		// Update database paid status
		// Find user from database
		db := h.DB.Clone()
		defer db.Close()
		if err = db.DB("buzzme").C("posts").UpdateId(p.ID, p); err != nil {
			if err == mgo.ErrNotFound {
				return echo.ErrNotFound
			}
			return
		}

	}

	return c.Redirect(http.StatusMovedPermanently, "/post/"+p.ID.Hex())
}
