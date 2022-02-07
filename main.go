package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	jsonextract "github.com/blainemoser/JsonExtract"
	logging "github.com/blainemoser/Logging"
	"github.com/blainemoser/MySqlDB/database"
	utils "github.com/blainemoser/goutils"
	slackresponse "github.com/blainemoser/slackResponse"
	"github.com/blainemoser/todobot/api"
	"github.com/blainemoser/todobot/event"
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
	db     *database.Database
	logger *logging.Log
	a      *api.Api
	env    map[string]string
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
	go processQueue(a)
	<-hold
}

func processQueue(a *api.Api) {
	tick := time.NewTicker(time.Second * 5)
	var err error
	for range tick.C {
		c := make(chan []map[string]string, 1)
		event.ProcessQueue(c)
		result := <-c
		if len(result) < 1 {
			continue
		}
		err = queueResult(result)
		if err != nil {
			a.ErrLog(err, false)
		}
	}
}

func queueResult(result []map[string]string) error {
	var err error
	errs := make([]string, 0)
	for _, v := range result {
		if v["heading"] == "" || v["message"] == "" {
			continue
		}
		err = slackresponse.SlackPost(v["heading"], v["message"], "INFO", a.SlackURL, logger)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}
	return nil
}

func bootstrap() {
	var err error
	logger = getLogger()
	db, err = getDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = parseEnv()
	if err != nil {
		log.Fatal(err)
	}
	a = api.Boot(getPort(), env["slackURL"], env["slackToken"], db, logger)
	err = event.BootQueue(db)
	if err != nil {
		log.Fatal(err)
	}
}

func parseEnv() error {
	env = make(map[string]string)
	envConfigs, err := utils.FileConfigs("env.json")
	if err != nil {
		return err
	}
	err = slackURL(envConfigs)
	if err != nil {
		return err
	}
	return slackToken(envConfigs)
}

func slackURL(envConfigs jsonextract.JSONExtract) error {
	slackURLInterface, err := envConfigs.Extract("slackURL")
	if err != nil {
		return err
	}
	slackURL := utils.StringInterface(slackURLInterface)
	if len(slackURL) < 1 {
		return fmt.Errorf("no slack url found")
	}
	env["slackURL"] = slackURL
	return nil
}

func slackToken(envConfigs jsonextract.JSONExtract) error {
	slackTokenInterface, err := envConfigs.Extract("slackToken")
	if err != nil {
		return err
	}
	slackToken := utils.StringInterface(slackTokenInterface)
	if len(slackToken) < 1 {
		return fmt.Errorf("no slack token found")
	}
	env["slackToken"] = slackToken
	return nil
}

func getDatabase() (*database.Database, error) {
	configs, err := utils.FileConfigs("./migrations/configs.json")
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
