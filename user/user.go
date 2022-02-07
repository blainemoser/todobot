package user

import (
	"fmt"
	"strings"

	jsonextract "github.com/blainemoser/JsonExtract"
	"github.com/blainemoser/MySqlDB/database"
	utils "github.com/blainemoser/goutils"
)

const newUser = `insert into users (uhash, tz, tz_label, tz_offset) values(?,?,?,?)`

const editUser = `update users set tz = ?, tz_label = ?, tz_offset = ? where id = ?`

const findUser = `select * from users where id = ?`

const getAll = `select * from users`

type User struct {
	*database.Database
	uhash    string
	id       int64
	tzOffset int64
	tzLabel  string
	tz       string
}

type UserInit struct {
	*database.Database
	Uhash    string
	TZOffset int64
	TZLabel  string
	TZ       string
}

func UsersList(db *database.Database) (map[int64]*User, error) {
	result := make(map[int64]*User)
	records, err := db.QueryRaw(getAll, nil)
	if err != nil {
		return map[int64]*User{}, err
	}
	for _, urec := range records {
		addToUserList(db, urec, &result)
	}
	return result, nil
}

func Create(ui *UserInit) (*User, error) {
	u := &User{
		Database: ui.Database,
		uhash:    ui.Uhash,
		tzOffset: ui.TZOffset,
		tzLabel:  ui.TZLabel,
		tz:       ui.TZ,
	}
	result, err := u.Exec(newUser, []interface{}{
		u.uhash, u.tz, u.tzLabel, u.tzOffset,
	})
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
	u.tzOffset = utils.Int64Interface(result["tz_offset"])
	u.tz = utils.StringInterface(result["tz"])
	u.tzLabel = utils.StringInterface(result["tz_label"])
	return u, nil
}

func CreateFromRecord(r map[string]interface{}, db *database.Database) (*User, error) {
	uhash := utils.StringInterface(r["uhash"])
	id := utils.Int64Interface(r["id"])
	tzOffset := utils.Int64Interface(r["tz_offset"])
	tz := utils.StringInterface(r["tz"])
	tzLabel := utils.StringInterface(r["tz_label"])
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
		tzOffset: tzOffset,
		tz:       tz,
		tzLabel:  tzLabel,
	}, nil
}

func NewInit(db *database.Database, rawPayload string, uhash string) (*UserInit, error) {
	ext := jsonextract.JSONExtract{
		RawJSON: rawPayload,
	}
	errs := make([]string, 0)
	result := make(map[string]interface{})
	for _, v := range []string{"user/tz", "user/tz_label", "user/tz_offset"} {
		tzDetailsResult(ext, v, &result, &errs)
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf(strings.Join(errs, ", "))
	}
	return makeInit(result, uhash, db)
}

func (u *User) Update() (*User, error) {
	_, err := u.Exec(editUser, []interface{}{
		u.tz, u.tzLabel, u.tzOffset, u.id,
	})
	return u, err
}

func (u *User) UpdateFromInit(ui *UserInit) *User {
	u.tz = ui.TZ
	u.tzLabel = ui.TZLabel
	u.tzOffset = ui.TZOffset
	return u
}

func (u *User) SetTZ(tz string) *User {
	u.tz = tz
	return u
}

func (u *User) SetTZOffset(tzOffset int64) *User {
	u.tzOffset = tzOffset
	return u
}

func (u *User) SetTZLabel(tzLabel string) *User {
	u.tzLabel = tzLabel
	return u
}

func (u *User) Hash() string {
	return u.uhash
}

func (u *User) ID() int64 {
	return u.id
}

func (u *User) TZOffset() int64 {
	return u.tzOffset
}

func (u *User) TZ() string {
	return u.tz
}

func tzDetailsResult(ext jsonextract.JSONExtract, v string, result *map[string]interface{}, errs *[]string) {
	resInterface, err := ext.Extract(v)
	if err != nil {
		*errs = append(*errs, err.Error())
		return
	}
	(*result)[strings.Replace(v, "user/", "", 1)] = resInterface
}

func makeInit(result map[string]interface{}, uhash string, db *database.Database) (*UserInit, error) {
	tz := utils.StringInterface(result["tz"])
	tzOffset := int64(utils.Float64Interface(result["tz_offset"]))
	tzLabel := utils.StringInterface(result["tz_label"])
	if tzOffset <= 0 {
		return nil, fmt.Errorf("timezone offset not found")
	}
	return &UserInit{
		Database: db,
		Uhash:    uhash,
		TZOffset: tzOffset,
		TZLabel:  tzLabel,
		TZ:       tz,
	}, nil
}

func addToUserList(db *database.Database, urec map[string]interface{}, result *map[int64]*User) {
	id := utils.Int64Interface(urec["id"])
	tzOffset := utils.Int64Interface(urec["tz_offset"])
	tz := utils.StringInterface(urec["tz"])
	tzLabel := utils.StringInterface(urec["tz_label"])
	u := &User{
		Database: db,
		uhash:    utils.StringInterface(urec["uhash"]),
		id:       id,
		tzOffset: tzOffset,
		tzLabel:  tzLabel,
		tz:       tz,
	}
	(*result)[id] = u
}
