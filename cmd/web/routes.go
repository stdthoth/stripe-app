package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(LoadSessions)

	mux.Get("/", app.HomePage)

	mux.Get("/pos-terminal", app.PosTerminal)
	mux.Post("/pos-terminal-payment-succeded", app.PosPaymentSucceeded)
	mux.Get("/pos-payment-receipt", app.PosReceipt)

	mux.Get("/receipt", app.Receipt)
	mux.Post("/payment-succeded", app.PaymentSucceeded)
	mux.Get("/img/{id}", app.ChargeOnce)
	mux.Post("/subscriptions/gold", app.GoldPlan)

	fileserver := http.FileServer(http.Dir("./static"))
	mux.Handle("./static/*", http.StripPrefix("/static", fileserver))
	return mux
}
