package process

import (
	"github.com/blainemoser/MySqlDB/database"
	"github.com/blainemoser/todobot/api"
)

type Process struct {
	DB  *database.Database
	API *api.Api
}
