package user

import (
	"fmt"

	"github.com/blainemoser/MySqlDB/database"
	utils "github.com/blainemoser/goutils"
)

const newUser = `insert into users (uhash) values(?)`

const findUser = `select * from users where id = ?`

const getAll = `select * from users`

type User struct {
	*database.Database
	uhash string
	id    int64
}

func UsersList(db *database.Database) (map[int64]*User, error) {
	result := make(map[int64]*User)
	records, err := db.QueryRaw(getAll, nil)
	if err != nil {
		return map[int64]*User{}, err
	}
	var u *User
	var id int64
	for _, urec := range records {
		id = utils.Int64Interface(urec["id"])
		u = &User{
			Database: db,
			uhash:    utils.StringInterface(urec["uhash"]),
			id:       id,
		}
		result[id] = u
	}
	return result, nil
}

func Create(db *database.Database, uhash string) (*User, error) {
	u := &User{
		Database: db,
		uhash:    uhash,
	}
	result, err := u.Exec(newUser, []interface{}{u.uhash})
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	u.id = id
	return u, nil
}

func Find(db *database.Database, id int64) (*User, error) {
	result, err := db.Row(findUser, id)
	if err != nil {
		return nil, err
	}
	u := &User{
		id: id,
	}
	u.uhash = utils.StringInterface(result["uhash"])
	return u, nil
}

func CreateFromRecord(r map[string]interface{}, db *database.Database) (*User, error) {
	uhash := utils.StringInterface(r["uhash"])
	id := utils.Int64Interface(r["id"])
	if len(uhash) < 1 {
		return nil, fmt.Errorf("name not provided")
	}
	if id < 1 {
		return nil, fmt.Errorf("id not found")
	}
	return &User{
		Database: db,
		id:       id,
		uhash:    uhash,
	}, nil
}

func (u *User) Hash() string {
	return u.uhash
}

func (u *User) ID() int64 {
	return u.id
}
