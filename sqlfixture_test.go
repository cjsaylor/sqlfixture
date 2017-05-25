// Required environment variables:
//   * MYSQL_DSN - For example: username:password@tcp(localhost:3306)/test

package sqlfixture_test

import (
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cjsaylor/sqlfixture"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

var setupQueries = []string{
	"DROP TABLE IF EXISTS `test`",
	"DROP TABLE IF EXISTS `test2`",
	`CREATE TABLE test (
	id int(11) unsigned,
	name varchar(30),
	PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
	`,
	`CREATE TABLE test2 (
	id int(11) unsigned,
	slug varchar(30),
	PRIMARY KEY (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
	`,
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	dsn := os.Getenv("MYSQL_DSN")
	db, _ = sql.Open("mysql", dsn)
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	for _, q := range setupQueries {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

type test2Result struct {
	id   int64
	slug string
}

func TestFromYAML(t *testing.T) {
	filename, _ := filepath.Abs("./fixtures/test.yaml")
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Error("Unable to find fixture file.")
		return
	}
	fixture, err := sqlfixture.FromYAML(db, yamlFile)
	if err != nil {
		t.Error(err)
		return
	}
	fixture.Populate()
	evaluateCommonFixture(t)
}

func TestFromJSON(t *testing.T) {
	filename, _ := filepath.Abs("./fixtures/test.json")
	jsonFile, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Error("Unable to find fixture file.")
		return
	}
	fixture, err := sqlfixture.FromJSON(db, jsonFile)
	if err != nil {
		t.Error(err)
		return
	}
	fixture.Populate()
	evaluateCommonFixture(t)
}

func TestFixture_Populate(t *testing.T) {
	fixture := sqlfixture.New(db, sqlfixture.Tables{
		sqlfixture.Table{
			Name: "test",
			Rows: sqlfixture.Rows{
				sqlfixture.Row{
					"id":   "1",
					"name": "something",
				},
			},
		},
		sqlfixture.Table{
			Name: "test2",
			Rows: sqlfixture.Rows{
				sqlfixture.Row{
					"id":   "1",
					"slug": "something",
				},
				sqlfixture.Row{
					"id":   "2",
					"slug": "something-else",
				},
			},
		},
	})
	fixture.Populate()
	evaluateCommonFixture(t)
}

func evaluateCommonFixture(t *testing.T) {
	var id int64
	var name string
	err := db.QueryRow("select id, name from test").Scan(&id, &name)
	if err != nil {
		t.Error(err)
		return
	}
	if id != 1 || name != "something" {
		t.Errorf("Row check failed.")
		return
	}
	var results []test2Result
	rows, err := db.Query("select id, slug from test2")
	defer rows.Close()
	for rows.Next() {
		result := test2Result{}
		rows.Scan(result.id, result.slug)
		results = append(results, result)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %v", len(results))
	}
}
