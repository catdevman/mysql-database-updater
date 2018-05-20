package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yunabe/easycsv"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	dbUsername, dbPassword, dbHost, defaultDatabase, environment, sqlFile, environmentsFile, databasePrefix string
)

func main() {
	flag.StringVar(&environment, "env", "local", "Choose environment from environments file")
	flag.StringVar(&environmentsFile, "envFile", "environments.csv", "Choose path for environments file")
	flag.StringVar(&sqlFile, "sqlFile", "updates.sql", "Choose path for sql file")
	flag.StringVar(&databasePrefix, "dbPrefix", "db_", "Choose a prefix for the databases that this will loop over")
	flag.Parse()
	dbUsername, dbPassword, dbHost = getDBConnectionParameters(environmentsFile, environment)
	db, err := getDatabaseConnection(defaultDatabase)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	sqlStr, err := getSQLContents(sqlFile)
	if err != nil {
		log.Println(err)
		os.Exit(3)
	}

	databases, err := getDatabases()
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}
	dbs := make(chan string)
	go func() {
		for _, database := range databases {
			dbs <- database
		}
		close(dbs)
	}()

	r1 := processSQL(dbs, sqlStr)
	r2 := processSQL(dbs, sqlStr)
	r3 := processSQL(dbs, sqlStr)
	r4 := processSQL(dbs, sqlStr)
	r5 := processSQL(dbs, sqlStr)
	r6 := processSQL(dbs, sqlStr)
	r7 := processSQL(dbs, sqlStr)
	r8 := processSQL(dbs, sqlStr)

	for val := range merge(r1, r2, r3, r4, r5, r6, r7, r8) {
		log.Println(val)
	}
}

func getSQLContents(filename string) (string, error) {
	bs, err := ioutil.ReadFile(filename)
	return string(bs), err
}

func getDatabaseConnection(database string) (*sql.DB, error) {
	return sql.Open("mysql", buildDsnString(dbUsername, dbPassword, dbHost, database))
}

func buildDsnString(user string, password, host string, database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?multiStatements=true", user, password, host, database)
}

func getDBConnectionParameters(filename string, environment string) (string, string, string) {
	var u, p, h string
	r := easycsv.NewReaderFile(filename)
	var header struct {
		Environment string `name:"environment"`
		Username    string `name:"username"`
		Password    string `name:"password"`
		Host        string `name:"host"`
	}

	for r.Read(&header) {
		if header.Environment == environment {
			u = header.Username
			p = header.Password
			h = header.Host
		}
	}
	return u, p, h
}

func getDatabases() ([]string, error) {
	var databases []string
	db, err := getDatabaseConnection(defaultDatabase)
	if err != nil {
		return databases, err
	}
	rows, err := db.Query("SHOW DATABASES;")
	if err != nil {
		return databases, err
	}
	defer rows.Close()
	var d string
	for rows.Next() {
		rows.Scan(&d)
		if strings.HasPrefix(d, databasePrefix) {
			databases = append(databases, d)
		}
	}

	err = rows.Err()

	return databases, err
}

func runSQL(database string, sqlStr string) error {
	db, err := getDatabaseConnection(database)
	if err != nil {
		log.Println(err)
	}

	_, err = db.Exec(sqlStr)
	db.Close()
	return err
}

func processSQL(ch chan string, sqlStr string) chan string {
	out := make(chan string)
	go func() {
		for database := range ch {
			output := "Completed with: " + database
			err := runSQL(database, sqlStr)
			if err != nil {
				output = err.Error()
			}
			out <- output
		}
		close(out)
	}()

	return out
}

func merge(cs ...chan string) chan string {
	out := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(cs))

	for _, c := range cs {
		go func(ch chan string) {
			for s := range ch {
				out <- s
			}
			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
