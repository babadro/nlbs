package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/babadro/enlbs/repo"
	"github.com/valyala/fasthttp"
	"log"
)

type req struct {
	State         string  `json:"state"`
	Amount        float32 `json:"amount"`
	TransactionId int64   `json:"transactionId"`
}

func main() {
	flag.UintVar(&AppConfig.Port, "port", 8888, "port")
	flag.StringVar(&AppConfig.DBUser, "dbuser", "postgres", "db user")
	flag.StringVar(&AppConfig.DBName, "dbname", "test", "db name")
	flag.StringVar(&AppConfig.DBPass, "dbpass", "postgrespass", "db password")
	flag.BoolVar(&AppConfig.CleanDB, "clean", true, "clean db")
	flag.Parse()

	db, err := repo.OpenDB(AppConfig.DBUser, AppConfig.DBName, AppConfig.DBPass)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	if AppConfig.CleanDB {
		_, _ = db.Exec("DELETE FROM transactions")
		_, _ = db.Exec("DELETE FROM users")
		_, _ = db.Exec("INSERT INTO users (balance, id) VALUES (0, 1)")
	}

	httpServer := fasthttp.Server{}
	httpServer.Handler = createRequestHandler(db)

	addr := fmt.Sprintf(":%d", AppConfig.Port)

	log.Println("API started", addr)
	log.Fatal(httpServer.ListenAndServe(addr))
}

func createRequestHandler(db *sql.DB) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch path {
		case "/balance-update":
			if string(ctx.Request.Header.Method()) != "POST" {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				return
			}
			if err := processRequest(ctx, db); err != nil {
				ctx.SetBodyString(err.Error())
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				return
			}
			ctx.SetStatusCode(fasthttp.StatusOK)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}
}

func processRequest(ctx *fasthttp.RequestCtx, db *sql.DB) error {
	var r req
	if err := json.Unmarshal(ctx.Request.Body(), &r); err != nil {
		return err
	}
	if r.Amount <= 0 {
		return errors.New("Amount must be a positive number")
	}
	if r.TransactionId <= 0 {
		return errors.New("TransactionId must be a positive number")
	}
	var amount float32
	switch r.State {
	case "win":
		amount = r.Amount
	case "lost":
		amount = -r.Amount
	default:
		return errors.New("unknown State")
	}
	if _, err := db.Exec(repo.Statement, amount, r.TransactionId); err != nil {
		return err
	}
	return nil
}
