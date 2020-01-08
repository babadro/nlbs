package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"github.com/babadro/enlbs/repo"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	win  = "win"
	lost = "lost"
)

type config struct {
	mode    string
	dbUser  string
	dbName  string
	dbPass  string
	cleanDb bool
	port    int
}

var conf config

func main() {
	flag.StringVar(&conf.mode, "mode", "api", "mode")
	flag.BoolVar(&conf.cleanDb, "clean", true, "clean db")
	flag.IntVar(&conf.port, "port", 8888, "port")
	flag.StringVar(&conf.dbUser, "dbuser", "postgres", "db user")
	flag.StringVar(&conf.dbName, "dbname", "test", "db name")
	flag.StringVar(&conf.dbPass, "dbpass", "postgrespass", "db password")
	flag.Parse()

	db, err := repo.OpenDB(conf.dbUser, conf.dbName, conf.dbPass)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	if conf.cleanDb {
		_, _ = db.Exec("DELETE FROM transactions")
		_, _ = db.Exec("DELETE FROM users")
		_, _ = db.Exec("INSERT INTO users (balance, id) VALUES (0, 1)")
	}

	switch conf.mode {
	case "sql":
		for i := 0; i < 10; i++ {
			go updateLoop(db, update)
			time.Sleep(10 * time.Millisecond)
		}
	case "api":
		go updateLoop(db, updateByApi)
	case "sync":
		go syncLoop(db)
	default:
		log.Fatal("unknown mode")
	}
	go cancelLoop(db)
	time.Sleep(3 * time.Second)
}

func syncLoop(db *sql.DB) {
	rnd := rand.New(rand.NewSource(1))
	i := 0
	for {
		update(db, rnd)
		if i%15 == 0 {
			cancel(db)
		}
		i++
	}
}

func updateLoop(db *sql.DB, upd func(*sql.DB, *rand.Rand)) {
	rnd := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	for {
		upd(db, rnd)
		time.Sleep(10 * time.Millisecond)
	}
}

func update(db *sql.DB, rnd *rand.Rand) {
	id := rnd.Int63()
	_, err := db.Exec(repo.Statement, randAmount(rnd), id)
	if err != nil {
		log.Println("id= ", id)
		log.Fatal(err)
	}
}

func updateByApi(db *sql.DB, rnd *rand.Rand) {
	id := rnd.Int63()
	amount := randAmount(rnd)
	var state string
	if amount < 0 {
		state = lost
		amount = -amount
	} else {
		state = win
	}

	url := fmt.Sprintf("http://localhost:%d/balance-update", conf.port)
	var payload = []byte(fmt.Sprintf(`{"state":%q,"amount":%f,"transactionId":%d}`, state, amount, id))
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	if err != nil {
		log.Println(url)
		log.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		log.Println(url)
		body, _ := ioutil.ReadAll(response.Body)
		log.Fatalf("api response status %d, body: %s", response.StatusCode, body)
	}
}

func cancelLoop(db *sql.DB) {
	for {
		cancel(db)
		time.Sleep(10 * time.Millisecond)
	}
}

func cancel(db *sql.DB) {
	_, err := db.Exec("SELECT transactions_cancel()")
	if err != nil {
		log.Println("cancel err")
		log.Fatal(err)
	}
}

func randAmount(rand *rand.Rand) float32 {
	res := float32(rand.Intn(100))
	signed := false
	if rand.Intn(2) == 1 {
		signed = true
	}
	res += float32(rand.Intn(100)) / 100
	if signed {
		res = -res
	}
	return res
}
