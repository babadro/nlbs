// Just for test
package main

import (
	"flag"
	"github.com/babadro/enlbs/repo"
	"log"
	"time"
)

type config struct {
	timeout time.Duration
	dbUser  string
	dbName  string
	dbPass  string
}

func main() {
	var conf config
	flag.DurationVar(&conf.timeout, "timeout", 60*time.Second, "postprocessing timeout")
	flag.StringVar(&conf.dbUser, "dbuser", "postgres", "db user")
	flag.StringVar(&conf.dbName, "dbname", "test", "db name")
	flag.StringVar(&conf.dbPass, "dbpass", "postgrespass", "db password")
	flag.Parse()

	db, err := repo.OpenDB(conf.dbUser, conf.dbName, conf.dbPass)
	if err != nil {
		log.Fatal(err)
	}
	for {
		_, err := db.Exec("SELECT transactions_cancel()")
		if err != nil {
			log.Println("transactions_cancel: ", err)
		}
		time.Sleep(conf.timeout)
	}
}
