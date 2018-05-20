package main

import (
	"testing"
)

const dsnStringTest = "root:password@tcp(127.0.0.1:3306)/pivot2_apigiliy?multiStatements=true"

func TestGetDBConnectionParameters(t *testing.T) {
	dbUsername1, dbPassword1, dbHost1 := getDBConnectionParameters("testdata/environments.csv", "local")
	if dbUsername1 != "root" || dbPassword1 != "password" || dbHost1 != "127.0.0.1:3306" {
		t.Fail()
	}

	dbUsername2, dbPassword2, dbHost2 := getDBConnectionParameters("testdata/environments.csv", "test")
	if dbUsername2 != "root" || dbPassword2 != "password" || dbHost2 != "127.0.0.1:3306" {
		t.Fail()
	}

	dbUsername3, dbPassword3, dbHost3 := getDBConnectionParameters("testdata/environments.csv", "fail")
	if dbUsername3 != "fail" || dbPassword3 != "fail" || dbHost3 != "fail:3306" {
		t.Fail()
	}

	dbUsername4, dbPassword4, dbHost4 := getDBConnectionParameters("testdata/environments.csv", "bob")
	if dbUsername4 != "" || dbPassword4 != "" || dbHost4 != "" {
		t.Fail()
	}

	dbUsername5, dbPassword5, dbHost5 := getDBConnectionParameters("testdata/fail.fail", "fail")
	if dbUsername5 != "" || dbPassword5 != "" || dbHost5 != "" {
		t.Fail()
	}
}

func TestBuildDsnString(t *testing.T) {
	databasePrefix = "test_"
	defaultDatabase = "test_database"
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
	for _ = range merge(in1, in2) {
		count++
	}

	if count != 4 {
		t.Fail()
	}
}

func TestRunSQL(t *testing.T) {
	databasePrefix = "test_"
	defaultDatabase = "test_database"
	err := runSQL("fail", "SELECT fail from dual;")
	if err == nil {
		t.Fail()
	}
}

func TestProcessSQL(t *testing.T) {
	databasePrefix = "test_"
	defaultDatabase = "test_database"
	databases := make(chan string)
	go func() {
		databases <- "fail1"
		databases <- "fail2"
		close(databases)
	}()

	c := processSQL(databases, "select fail from dual;")

	count := 0
	for _ = range c {
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
	// defaultDatabase = "pivot_all"
	// dbUsername, dbPassword, dbHost = getDBConnectionParameters("testdata/environments.csv", "fail")
	// println(dbUsername, dbPassword, dbHost)
	// _, err := getDatabases()
	// if err == nil {
	// 	t.Fail()
	// }
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
