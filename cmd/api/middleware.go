package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) rateLimit(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorRespone(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorRespone(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func (app *application) enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        allowedOrigins := map[string]bool{
            "http://localhost:5173": true,
            "https://mayura-andrew.github.io": true,
			"https://mayura-andrew.github.io/": true,
			"https://mayura-andrew.github.io/knowlihub": true,
			"https://mayura-andrew.github.io/knowlihub/": true,
			"http://mayura-andrew.github.io/knowlihub": true,
			"http://mayura-andrew.github.io/knowlihub/": true,
			"http://mayura-andrew.github.io": true,
			"http://mayura-andrew.github.io/": true,
            "*": true,

        }

        origin := r.Header.Get("Origin")

        if _, ok := allowedOrigins[origin]; ok {
            w.Header().Set("Access-Control-Allow-Origin", origin)
        } else if allowedOrigins["*"] {
            w.Header().Set("Access-Control-Allow-Origin", "*")
        } else {
            log.Printf("CORS not allowed for origin: %s\n", origin)
            return
        }

        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
        log.Printf("CORS allowed for origin: %s\n", origin)

        next.ServeHTTP(w, r)
    })
}
