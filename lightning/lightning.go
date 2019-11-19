package lightning

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type AddressResponse struct {
	Address string `json:'address'`
}

// Set admin.macaroon hex
const (
	Macaroon = "0201036C6E6402CF01030A1055FDF8595BEFE82D5A695EAFEBA8462E1201301A160A0761646472657373120472656164120577726974651A130A04696E666F120472656164120577726974651A170A08696E766F69636573120472656164120577726974651A160A076D657373616765120472656164120577726974651A170A086F6666636861696E120472656164120577726974651A160A076F6E636861696E120472656164120577726974651A140A057065657273120472656164120577726974651A120A067369676E6572120867656E657261746500000620BB7928E866D9448A6C8FEB419D83DCD166E806E9CB3B9BA200282309567A2F8E"
	LNUrl    = "https://localhost:8080/v1/"
)

func sendGetRequest(endpoint string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", LNUrl+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", Macaroon)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func sendPostRequest(endpoint string, payload string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	log.Println(payload)

	/*params := bytes.NewBuffer(nil)
	if payload != nil {
		if err := json.NewEncoder(params).Encode(payload); err != nil {
			return nil, err
		}
	}*/
	var jsonStr = []byte(payload)

	req, err := http.NewRequest("POST", LNUrl+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", Macaroon)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type InvoiceRequest struct {
	cltv_expiry      string
	add_index        string
	creation_date    string
	private          bool
	value            string `json:"value"`
	expiry           string
	fallback_addr    string
	r_hash           byte
	memo             string `json:"memo"`
	receipt          byte
	amt_paid_msat    string
	payment_request  string
	description_hash byte
	settle_index     string
	settle_date      string
	settled          bool
	r_preimage       byte
	amt_paid_sat     string
}

type InvoiceResponse struct {
	RHash          byte   `json:"r_hash"`
	PaymentRequest string `json:"payment_request"`
	AddIndex       string `json:"add_index"`
}

type InvoiceListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
}

func CreateInvoice(amount string) (string, error) {
	log.Println(amount)

	resp, err := sendPostRequest("invoices", `{"value":"`+amount+`","memo":"`+amount+`"}`)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, err
}

func GetInvoicePaid(invoice InvoiceResponse) (bool, error) {
	var (
		invoiceValid   = false
		invoicePending = false
		invoicePaid    = false
	)

	// First see if invoice exists
	resp, err := sendGetRequest("invoices")
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	validInvoices := InvoiceListResponse{}
	json.Unmarshal(bodyBytes, &validInvoices)

	for _, validInvoice := range validInvoices.Invoices {
		if validInvoice.PaymentRequest == invoice.PaymentRequest {
			invoiceValid = true
			break
		}
	}

	resp, err = sendGetRequest("invoices?pending_only=true")
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	pendingInvoices := InvoiceListResponse{}
	json.Unmarshal(bodyBytes, &pendingInvoices)

	for _, pendingInvoice := range pendingInvoices.Invoices {
		if pendingInvoice.PaymentRequest == invoice.PaymentRequest {
			invoicePending = true
			break
		}
	}

	if invoiceValid && !invoicePending {
		invoicePaid = true
	}

	//TESTING
	invoicePaid = true

	return invoicePaid, err
}

func GetPaymentRequestValid(paymentRequest string) bool {
	// First see if invoice exists
	resp, err := sendGetRequest("payreq/" + paymentRequest)
	if err != nil {
		return false
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	bodyString := string(bodyBytes)

	if strings.Contains(bodyString, "err") {
		return false
	}

	return true
}
