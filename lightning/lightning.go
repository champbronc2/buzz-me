package lightning

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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

func sendPostRequest(endpoint string, payload interface{}) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", LNUrl+endpoint, bytes.NewBuffer(p))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", Macaroon)
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
	value            string
	expiry           string
	fallback_addr    string
	r_hash           byte
	memo             string
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

func CreateInvoice(amount string) (string, error) {
	newInvoiceRequest := InvoiceRequest{
		value: amount,
	}

	resp, err := sendPostRequest("invoices", newInvoiceRequest)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	log.Println(bodyString)

	return bodyString, err
}
