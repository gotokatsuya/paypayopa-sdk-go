package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gotokatsuya/paypayopa-sdk-go/paypay"
)

// envs
var (
	apiKey      = os.Getenv("PAYPAY_API_KEY")
	apiSecret   = os.Getenv("PAYPAY_API_SECRET")
	merchant    = os.Getenv("PAYPAY_MERCHANT")
	redirectURL = os.Getenv("PAYPAY_ACCOUNT_LINK_REDIRECT_URL")
)

func main() {
	paypayCli, err := paypay.New(apiKey, apiSecret, merchant, paypay.WithSandbox())
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp, httpResp, err := paypayCli.CreateAccountLinkQRCode(r.Context(), &paypay.CreateAccountLinkQRCodeRequest{
			Scopes: []string{
				"continuous_payments",
			},
			Nonce:        uuid.New().String(),
			RedirectType: "WEB_LINK",
			RedirectURL:  redirectURL,
			ReferenceID:  uuid.New().String(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		switch httpResp.StatusCode {
		case http.StatusCreated:
			http.Redirect(w, r, resp.Data.LinkQRCodeURL, http.StatusFound)
		default:
			http.Error(w, resp.ResultInfo.Code+" "+resp.ResultInfo.Message, http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/confirm", func(w http.ResponseWriter, r *http.Request) {
		responseToken := r.URL.Query().Get("responseToken")
		token, err := paypayCli.ParseResponseToken(responseToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("%v\n", token)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/recurring", func(w http.ResponseWriter, r *http.Request) {})

	http.HandleFunc("/recurring/close", func(w http.ResponseWriter, r *http.Request) {})

	fmt.Println("http://localhost:8000")
	log.Fatalln(http.ListenAndServe(":8000", nil))
}
