package main

import (
	_ "log"
	"runtime"
	"testing"
)

const dsnStringTest = "root:password@tcp(127.0.0.1:3306)/pivot2_apigiliy?multiStatements=true"

func TestGetDBConnectionParameters(t *testing.T) {
	dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/environments.csv", "local")
	if dbUsername != "root" || dbPassword != "password" || dbHost != "127.0.0.1:3306" {
		t.Fail()
	}

	dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/environments.csv", "test")
	if dbUsername != "root" || dbPassword != "" || dbHost != "127.0.0.1:3306" {
		t.Fail()
	}

	dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/environments.csv", "fail")
	if dbUsername != "fail" || dbPassword != "fail" || dbHost != "fail:3306" {
		t.Fail()
	}

	dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/environments.csv", "bob")
	if dbUsername != "" || dbPassword != "" || dbHost != "" {
		t.Fail()
	}

	dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/fail.fail", "fail")
	if dbUsername != "" || dbPassword != "" || dbHost != "" {
		t.Fail()
	}
}

func TestBuildDsnString(t *testing.T) {
	val := buildDsnString("root", "password", "127.0.0.1:3306", "pivot2_apigiliy")
	if val != dsnStringTest {
		t.Fail()
	}
}

func TestMerge(t *testing.T) {
	in1 := make(chan string)
	in2 := make(chan string)
	go func() {
		in1 <- "in1-test1"
		in1 <- "in1-test2"
		in2 <- "in2-test1"
		in2 <- "in2-test2"
		close(in1)
		close(in2)
	}()
	count := 0
	for range merge(in1, in2) {
		count++
	}

	if count != 4 {
		t.Fail()
	}
}

func TestRunMultiSQL(t *testing.T) {
	db, _ := getDatabaseConnection("fail")
	err := runMultiSQL("SELECT fail from dual; SELECT fail2 from dual;", db)
	if err == nil {
		t.Fail()
	}
}

func TestProcessSQL(t *testing.T) {
	databases := make(chan string)
	go func() {
		databases <- "fail1"
		databases <- "fail2"
		close(databases)
	}()
	db, _ := getDatabaseConnection("fail")
	c := processSQL(databases, "select fail from dual;", db)

	count := 0
	for range c {
		count++
	}

	if count != 2 {
		t.Fail()
	}
}

func TestGetSQLContents(t *testing.T) {
	sqlStr, err := getSQLContents("testdata/test.sql")
	if err != nil {
		t.Fail()
	}
	if sqlStr != "SELECT now() FROM dual;\n" {
		t.Fail()
	}
}
func TestGetDatabases(t *testing.T) {
	dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/environments.csv", "fail")
	_, err := getDatabases()
	if err == nil {
		t.Fail()
	}

	dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/environments.csv", "test")
	databases, _ := getDatabases()

	if len(databases) == 0 {
		t.Fail()
	}
}

func TestGetDatabaseConnection(t *testing.T) {
	db, err := getDatabaseConnection("fail")
	if err != nil {
		t.Fail()
	}
	err = db.Ping()
	if err == nil {
		t.Fail()
	}
}

func TestGetDatabasesChannel(t *testing.T) {
	databases := []string{"test1", "test2"}
	dbs := getDatabasesChannel(databases)
	c := 0
	for range dbs {
		c++
	}
	if c != 2 {
		t.Fail()
	}
}

func TestCreateWorkers(t *testing.T) {
	c := 0
	db, _ := getDatabaseConnection("fail")
	defer db.Close()
	ch := make(chan string)
	results := createWorkers(ch, "select * from dual", db)
	for range results {
		c++
	}
	if c != runtime.NumCPU() {
		t.Fail()
	}
}
