package main

import (
	"expvar"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthcheckHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	router.HandlerFunc(http.MethodGet, "/", app.rootHandler)

	router.HandlerFunc(http.MethodPost, "/api/v1/send", app.sendEmailHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/subscribe", app.subscribeHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/sent", app.showEmailHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/redirect", app.track)
	router.HandlerFunc(http.MethodGet, "/api/v1/recipients/:email", app.showEmailHandler)

	return app.enableCORS(app.recoverPanic(app.rateLimit(router)))
}
