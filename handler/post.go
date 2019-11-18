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
		Read: false,
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

	/*invoice, err := lightning.CreateInvoice(u.FeeRate)
	if err != nil {
		return err
	}

	p.Invoice = invoice*/
	p.Invoice = "{\"r_hash\":\"rGZUpNuwPGdysrfIEf7iso4PlrJHDUlIRCNsyDDO1E0=\",\"payment_request\":\"lntb1pwukphgpp543n9ffxmkq7xwu4jklyprlhzk28ql94jgux5jjzyydkvsvxw63xsdqqcqzpgraqptwft5jhckznertz77nu0zh4vd9afgwlyr352z4u8gty73mk88qruxk5nt6a33pn6l2prvgp9mq503kz5rjluh9afkzyw747vztcqd7t9qp\",\"add_index\":\"1\"}"
	p.Sats = u.FeeRate

	invoice := lightning.InvoiceResponse{}
	json.Unmarshal([]byte(p.Invoice), &invoice)

	// Save post in database
	if err = db.DB("buzzme").C("posts").Insert(p); err != nil {
		return
	}

	return c.Render(http.StatusCreated, "post.html", map[string]interface{}{
		"id":             p.ID,
		"from":           p.From,
		"to":             p.To,
		"paymentRequest": invoice.PaymentRequest,
		"message":        p.Message,
		"paid":           p.Paid,
	})
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

	// Retrieve user from database
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
		"id":             p.ID,
		"from":           p.From,
		"to":             p.To,
		"paymentRequest": invoice.PaymentRequest,
		"message":        p.Message,
		"paid":           p.Paid,
	})
}
