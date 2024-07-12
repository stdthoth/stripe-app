package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ichthoth/stripe-credit-terminal/internal/cards"
	"github.com/ichthoth/stripe-credit-terminal/internal/models"
)

type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

func (app *application) PosTerminal(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplates(w, r, "terminal", &templateData{}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) HomePage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplates(w, r, "terminal", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var transaction TransactionData
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return transaction, err
	}

	//todo: add validation of fields
	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("email")
	paymentintent := r.Form.Get("payment_intent")
	paymentcurrency := r.Form.Get("payment_currency")
	paymentamount := r.Form.Get("payment_amount")
	paymentmethod := r.Form.Get("payment_method")
	amount, _ := strconv.Atoi(paymentamount)

	card := cards.Card{
		Secret: app.config.stripeInfo.secret,
		Key:    app.config.stripeInfo.key,
	}
	pi, err := card.RetrievePaymentIntent(paymentintent)
	if err != nil {
		app.errorLog.Println(err)
		return transaction, err
	}

	pm, err := card.Getpaymentmethod(paymentmethod)
	if err != nil {
		app.errorLog.Println(err)
		return transaction, err
	}

	lastFour := pm.Card.Last4
	cardExpiryM := pm.Card.ExpMonth
	cardExpiryY := pm.Card.ExpYear

	transaction = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		PaymentIntentID: paymentintent,
		PaymentMethodID: paymentmethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentcurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(cardExpiryM),
		ExpiryYear:      int(cardExpiryY),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}

	return transaction, nil
}

func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	gophimg, _ := strconv.Atoi(r.Form.Get("product_id"))

	transactionData, err := app.GetTransactionData(r)

	// creates the customer
	customerID, err := app.SaveCustomer(transactionData.FirstName, transactionData.LastName, transactionData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Println(customerID)

	//creates a new transaction
	if err != nil {
		app.errorLog.Println(err)
	}

	txn := models.Transaction{
		Amount:              transactionData.PaymentAmount,
		Currency:            transactionData.PaymentCurrency,
		LastFour:            transactionData.LastFour,
		BankReturnCode:      transactionData.BankReturnCode,
		ExpiryMonth:         transactionData.ExpiryMonth,
		ExpiryYear:          transactionData.ExpiryYear,
		PaymentIntent:       transactionData.PaymentIntentID,
		PaymentMethod:       transactionData.PaymentMethodID,
		TransactionStatusID: 2,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//creating a new order
	ord := models.Order{
		ImageID:       gophimg,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        transactionData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = app.SaveOrder(ord)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.Session.Put(r.Context(), "receipt", transactionData)
	http.Redirect(w, r, "/receipt", http.StatusSeeOther)
}

func (app *application) PosPaymentSucceeded(w http.ResponseWriter, r *http.Request) {

	transactionData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
	}

	txn := models.Transaction{
		Amount:              transactionData.PaymentAmount,
		Currency:            transactionData.PaymentCurrency,
		LastFour:            transactionData.LastFour,
		BankReturnCode:      transactionData.BankReturnCode,
		ExpiryMonth:         transactionData.ExpiryMonth,
		ExpiryYear:          transactionData.ExpiryYear,
		PaymentIntent:       transactionData.PaymentIntentID,
		PaymentMethod:       transactionData.PaymentMethodID,
		TransactionStatusID: 2,
	}

	_, err = app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.Session.Put(r.Context(), "receipt", transactionData)
	http.Redirect(w, r, "/pos-payment-receipt", http.StatusSeeOther)
}

func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn
	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplates(w, r, "suceeded", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) PosReceipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn
	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplates(w, r, "pos-receipt", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return id, nil

}

func (app *application) SaveTransaction(txn models.Transaction) (int, error) {
	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (app *application) SaveOrder(ord models.Order) (int, error) {
	id, err := app.DB.InsertOrder(ord)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	imageId, _ := strconv.Atoi(id)

	image, err := app.DB.GetGopherImages(imageId)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	data := make(map[string]interface{})
	data["images"] = image
	if err := app.renderTemplates(w, r, "buy", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

//subscriptions

func (app *application) GoldPlan(w http.ResponseWriter, r *http.Request) {
	images, err := app.DB.GetGopherImages(2)
	if err != nil {
		app.errorLog.Println(err)
	}
	data := make(map[string]interface{})
	data["images"] = images
	if err := app.renderTemplates(w, r, "gold-plan", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}

}
