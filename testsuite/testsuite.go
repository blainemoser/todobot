package testsuite

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	jsonextract "github.com/blainemoser/JsonExtract"
	logging "github.com/blainemoser/Logging"
	"github.com/blainemoser/MySqlDB/database"
	"github.com/blainemoser/MySqlMigrate/migrate"
	"github.com/blainemoser/TrySql/trysql"
	utils "github.com/blainemoser/goutils"
)

const timeFormat = "2006-01-02 15:04:05"

type TestSuite struct {
	TS           *trysql.TrySql
	TestDatabase *database.Database
	Database     string
	ResultCode   int
}

func Initialize(context string) (*TestSuite, error) {
	ts := &TestSuite{
		Database:   "todobot_test_database",
		ResultCode: 0,
	}
	trySql, err := trysql.Initialise([]string{"-v", "latest"})
	if err != nil {
		return nil, err
	}
	ts.TS = trySql
	err = ts.bootstrap(context)
	if err != nil {
		return nil, err
	}
	return ts, err
}

func (ts *TestSuite) TearDown() {
	fmt.Println("tearing down ... ")
	if ts != nil && ts.TestDatabase != nil {
		ts.TestDatabase.Close()
	}
	td := func(r interface{}) {
		if ts != nil && ts.TS != nil {
			err := ts.TS.TearDown()
			if err != nil {
				log.Println(err.Error())
			}
		}
		if r == nil {
			os.Exit(ts.ResultCode)
		}
		panic(r)
	}
	r := recover()
	td(r)
}

func (ts *TestSuite) bootstrap(context string) error {
	configs := ts.getConfigs(true)
	d, err := database.MakeSchemaless(configs)
	if err != nil {
		return err
	}
	err = ts.createSchema(&d)
	if err != nil {
		return err
	}
	baseDir, err := utils.BaseDir([]string{context}, "todobot")
	if err != nil {
		return err
	}
	return migrate.Make(ts.TestDatabase, fmt.Sprintf("%s/migrations", baseDir)).MigrateUp()
}

func (ts *TestSuite) getConfigs(schemaless bool) *database.Configs {
	var db string
	if schemaless {
		db = ""
	} else {
		db = ts.Database
	}
	return &database.Configs{
		Port:     ts.TS.HostPortStr(),
		Host:     "127.0.0.1",
		Username: "root",
		Password: ts.TS.Password(),
		Database: db,
		Driver:   "mysql",
	}
}

func (ts *TestSuite) createSchema(d *database.Database) error {
	_, err := d.Exec(fmt.Sprintf("create schema %s", ts.Database), nil)
	if err != nil {
		return err
	}
	d.Close()
	d.SetSchema(ts.Database)
	ts.TestDatabase = d
	return nil
}

func collateErrors(errs []string) error {
	if len(errs) < 1 {
		return nil
	}
	return fmt.Errorf(strings.Join(errs, ", "))
}

func (ts *TestSuite) setEnvVars() {
	os.Setenv("DB_PORT", ts.TS.HostPortStr())
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_USERNAME", "root")
	os.Setenv("DB_PASSWORD", ts.TS.Password())
	os.Setenv("DB_CONNECTION", "mysql")
}

func TestLogger() (*logging.Log, error) {
	baseDir, err := utils.BaseDir([]string{"testsuite"}, "todobot")
	if err != nil {
		return nil, err
	}
	return logging.NewLog(fmt.Sprintf("%s/%s", baseDir, "test_log.log"), "TESTING")
}

func EvaluateResult(result []byte, expects map[string]interface{}) (err error) {
	extract := jsonextract.JSONExtract{
		RawJSON: string(result),
	}
	errs := make([]string, 0)
	for path, expected := range expects {
		property, err := extract.Extract(path)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		if fmt.Sprintf("%v", property) != fmt.Sprintf("%v", expected) {
			errs = append(errs, fmt.Sprintf("expected %s to be %v, got %v", path, expected, property))
		}
	}
	return errStrings(errs)
}

func GetBody(r *http.Response) (data []byte, err error) {
	if r == nil || r.Body == nil {
		return nil, fmt.Errorf("nil response")
	}
	defer r.Body.Close()
	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(data) < 1 {
		return nil, fmt.Errorf("no response received")
	}
	return data, err
}

func errStrings(errs []string) (err error) {
	if errs == nil || len(errs) < 1 {
		return nil
	}
	return fmt.Errorf(strings.Join(errs, ", "))
}
