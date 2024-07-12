package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/stdthoth/stripe-app/internal/cards"
	"github.com/stripe/stripe-go/v72"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	LastFour      string `json:"last_four"`
	ExpiryMonth   string `json:"expiry_month"`
	ExpiryYear    string `json:"expiry_year"`
	Plan          string `json:"plan"`
	ProductId     string `json:"product_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type response struct {
	ID      int    `json:"id,omitempty"`
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		app.errorLog.Println(err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.errorLog.Println(err)
	}

	card := cards.Card{
		Currency: payload.Currency,
		Secret:   app.config.stripeInfo.secret,
		Key:      app.config.stripeInfo.key,
	}

	okay := true

	paymentIntent, msg, err := card.ChargeCard(payload.Currency, amount)
	if err != nil {
		okay = false
	}

	if okay {
		out, err := json.MarshalIndent(paymentIntent, " ", "")
		if err != nil {
			app.errorLog.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
	} else {
		resp := response{
			OK:      false,
			Message: msg,
			Content: "",
		}

		jsonOut, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			app.errorLog.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonOut)
	}

}

func (app *application) GetGopherImagesById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	imageId, _ := strconv.Atoi(id)

	image, err := app.DB.GetGopherImages(imageId)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	jsonOut, err := json.MarshalIndent(image, "", "  ")
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOut)

}

func (app *application) CreateCustomerAndSub(w http.ResponseWriter, r *http.Request) {
	var data stripePayload

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		app.errorLog.Println(err)
	}

	app.infoLog.Println(data.Amount, data.Currency, data.Email, data.PaymentMethod, data.Plan)

	card := cards.Card{
		Currency: data.Currency,
		Secret:   app.config.stripeInfo.secret,
		Key:      app.config.stripeInfo.key,
	}

	okay := true
	var subscription *stripe.Subscription

	stripeCustomer, msg, err := card.CreateCustomer(data.PaymentMethod, data.Email)
	if err != nil {
		app.errorLog.Println(err)
		okay = false
	}

	if okay {
		subscription, err = card.SubscribeToPlan(stripeCustomer, data.Plan, data.Email, data.LastFour, "")
		if err != nil {
			app.errorLog.Println(err)
			okay = false
		}

	}

	app.infoLog.Println("subscription is", subscription.ID)

	//msg = ""

	resp := response{
		OK:      okay,
		Message: msg,
	}

	jsonOut, err := json.MarshalIndent(resp, "", " ")
	if err != nil {
		app.errorLog.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonOut)
}
