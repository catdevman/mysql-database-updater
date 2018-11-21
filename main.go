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
	"runtime"
	"strings"
	"sync"
)

var (
	dbUsername, dbPassword, dbHost, defaultDatabase, environment, sqlFile, environmentsFile, databasePrefix, ignoreDatabasesFlag string
	ignoreDatabases                                                                                                              []string
)

func main() {
	flag.StringVar(&environmentsFile, "envFile", "environments.csv", "Choose path for environments file")
	flag.StringVar(&environment, "env", "local", "Choose environment from environments file")
	flag.StringVar(&sqlFile, "sqlFile", "updates.sql", "Choose path for sql file")
	flag.StringVar(&databasePrefix, "dbPrefix", "", "Choose a prefix for the databases that this will loop over (ex: db_)")
	flag.StringVar(&ignoreDatabasesFlag, "ignoredDbs", "", "Databases to ignore (ex: db1,db2 without prefix)")
	flag.Parse()
	dbUsername, dbPassword, dbHost = getDBConnectionParameters(environmentsFile, environment)
	ignoreDatabases = getDatabasesToIgnore(ignoreDatabasesFlag, databasePrefix)
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

	databases, err := getDatabases(ignoreDatabases)
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}

	dbs := getDatabasesChannel(databases)

	results := createWorkers(dbs, sqlStr, db)

	for val := range merge(results...) {
		log.Println(val)
	}
}

func createWorkers(ch chan string, sqlStr string, dbConn *sql.DB) (results []chan string) {
	for i := 0; i < runtime.NumCPU(); i++ {
		results = append(results, processSQL(ch, sqlStr, dbConn))
	}
	return
}

func getDatabasesChannel(databases []string) chan string {
	dbs := make(chan string)
	go func() {
		for _, database := range databases {
			dbs <- database
		}
		close(dbs)
	}()

	return dbs
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

func getDatabases(ignoredDatabases []string) ([]string, error) {
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
		if strings.HasPrefix(d, databasePrefix) && !contains(ignoredDatabases, d) {
			databases = append(databases, d)
		}
	}

	err = rows.Err()

	return databases, err
}

func runMultiSQL(sqlStr string, db *sql.DB) error {
	_, err := db.Exec(sqlStr)
	return err
}

func processSQL(ch chan string, sqlStr string, dbConn *sql.DB) chan string {
	out := make(chan string)
	go func(s string, db *sql.DB) {
		for database := range ch {
			output := "Completed with: " + database
			query := fmt.Sprintf("use `%s`; %s", database, s)
			err := runMultiSQL(query, db)
			if err != nil {
				output = err.Error()
			}
			out <- output
		}
		close(out)
	}(sqlStr, dbConn)

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

func getDatabasesToIgnore(ignoreDatabasesFlag, databasePrefix string) (ignoreDbs []string) {
	var dbs []string
	dbs = strings.Split(ignoreDatabasesFlag, ",")
	for _, d := range dbs {
		ignoreDbs = append(ignoreDbs, databasePrefix+d)
	}
	return ignoreDbs

}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
