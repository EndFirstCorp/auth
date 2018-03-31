package auth

import (
	"strconv"

	"github.com/EndFirstCorp/onedb"
	"github.com/pkg/errors"
)

type backendDbUser struct {
	Db onedb.DBer

	AddUserQuery              string
	GetUserQuery              string
	UpdateUserQuery           string
	CreateSecondaryEmailQuery string
}

// NewBackendDbUser creates a Postgres-based UserBackender
func NewBackendDbUser(server string, port int, username, password, database string, addUserQuery, getUserQuery, updateUserQuery, createSecondaryEmailQuery string) (UserBackender, error) {
	db, err := onedb.NewPgx(server, uint16(port), username, password, database)
	if err != nil {
		return nil, err
	}
	return &backendDbUser{Db: db,
		GetUserQuery:              getUserQuery,
		AddUserQuery:              addUserQuery,
		UpdateUserQuery:           updateUserQuery,
		CreateSecondaryEmailQuery: createSecondaryEmailQuery}, nil
}

func (u *backendDbUser) AddUser(email string) (string, error) {
	var userID int32 = -1
	return strconv.Itoa(int(userID)), u.Db.QueryValues(onedb.NewSqlQuery(u.AddUserQuery, email), &userID)
}

func (u *backendDbUser) GetUser(email string) (*User, error) {
	r := &User{}
	err := u.Db.QueryStructRow(onedb.NewSqlQuery(u.GetUserQuery, email), r)
	if err != nil {
		return nil, errors.New("Unable to get user: " + err.Error())
	}
	return r, err
}

func (u *backendDbUser) UpdateUser(userID string, fullname string, company string, pictureURL string) error {
	return u.Db.Execute(onedb.NewSqlQuery(u.UpdateUserQuery, userID, fullname))
}

func (u *backendDbUser) CreateSecondaryEmail(userID, secondaryEmail string) error {
	return u.Db.Execute(onedb.NewSqlQuery(u.CreateSecondaryEmailQuery, userID, secondaryEmail))
}

func (u *backendDbUser) Close() error {
	return nil
}
