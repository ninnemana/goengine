package database

import (
	"errors"
	"expvar"
	"github.com/ziutek/mymysql/mysql"
)

// prepared statements go here
var (
// example statement
//statementNameStmt = "select * from TableName"
)

// Create map of all statements
var (
	Statements map[string]mysql.Stmt
)

// Prepare all MySQL statements
func PrepareAll() error {

	Statements = make(map[string]mysql.Stmt, 0)

	if !Db.IsConnected() {
		Db.Connect()
	}

	// Example Preparation
	/*
		statementNamePrepared, err := Db.Prepare(statementNameStmt)
		if err != nil {
			return err
		}
		Statements["statementNameStmt"] = statementNamePrepared
	*/

	return nil
}

func GetStatement(key string) (stmt mysql.Stmt, err error) {
	stmt, ok := Statements[key]
	if !ok {
		qry := expvar.Get(key)
		if qry == nil {
			err = errors.New("Invalid query reference")
		} else {
			stmt, err = Db.Prepare(qry.String())
		}
	}
	return

}
