package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	jsonextract "github.com/blainemoser/JsonExtract"
	logging "github.com/blainemoser/Logging"
	"github.com/blainemoser/MySqlDB/database"
	utils "github.com/blainemoser/goutils"
	"github.com/blainemoser/todobot/api"
)

var (
	dbExpects = []string{
		"host",
		"port",
		"database",
		"username",
		"password",
		"driver",
	}
	db *database.Database
	a  *api.Api
)

func main() {
	hold := make(chan bool, 1)
	bootstrap()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go done(c)
	errc := make(chan error, 1)
	go func(errc chan error) {
		errc <- a.Run()
	}(errc)
	close(errc)
	err := <-errc
	if err != nil {
		log.Fatal(err)
	}
	<-hold
}

func bootstrap() {
	var err error
	logger := getLogger()
	db, err = getDatabase()
	if err != nil {
		log.Fatal(err)
	}
	a = api.Boot(getPort(), logger)
}

func getDatabase() (*database.Database, error) {
	configs, err := getDBConfigs()
	if err != nil {
		return nil, err
	}
	dbConfigs, err := extractDBConfigs(configs)
	if err != nil {
		return nil, err
	}
	db, err := database.Make(dbConfigs)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func extractDBConfigs(dbConfigs jsonextract.JSONExtract) (*database.Configs, error) {
	result := make(map[string]string)
	errs := make([]string, 0)
	for _, v := range dbExpects {
		config, err := dbConfigs.Extract(v)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		set, ok := config.(string)
		if !ok {
			errs = append(errs, fmt.Sprintf("%s is missing", v))
			continue
		}
		result[v] = set
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("missing configs: %s", strings.Join(errs, ", "))
	}
	return makeDBConfigs(result)
}

func makeDBConfigs(source map[string]string) (*database.Configs, error) {
	return &database.Configs{
		Host:     source["host"],
		Username: source["username"],
		Password: source["password"],
		Driver:   source["driver"],
		Database: source["database"],
		Port:     source["port"],
	}, nil
}

func getDBConfigs() (jsonextract.JSONExtract, error) {
	b, err := utils.GetFileContent("./migrations/configs.json")
	if err != nil {
		return jsonextract.JSONExtract{}, err
	}
	js := string(b)
	dbEnv := jsonextract.JSONExtract{
		RawJSON: js,
	}
	return dbEnv, nil
}

func done(signal chan os.Signal) {
	// wait on done signal, kill the process if received
	result := <-signal
	a.Write(result.String(), "INFO")
	os.Exit(1)
}

func getLogger() *logging.Log {
	baseDir, err := utils.BaseDir([]string{}, "todobot")
	if err != nil {
		log.Fatal(err)
	}
	l, err := logging.NewLog(fmt.Sprintf("%s/%s", baseDir, "log.log"), "PRODUCTION")
	if err != nil {
		log.Fatal(err)
	}
	return l
}

func getPort() int {
	args := os.Args
	var port string
	if len(args) > 1 {
		port = args[1]
	}
	if len(port) < 1 {
		port = "8081"
	}
	var err error
	result, err := strconv.ParseInt(port, 10, 24)
	if err != nil {
		log.Println(err)
		return 8081
	}
	return int(result)
}