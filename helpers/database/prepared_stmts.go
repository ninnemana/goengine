package database

import (
	"errors"
	"expvar"
	"github.com/ziutek/mymysql/mysql"
	_ "log"
)

var (
	Statements = make(map[string]mysql.Stmt, 0)
)

// Prepare all MySQL statements
func PrepareAll() error {

	UnPreparedStatements := make(map[string]string, 0)
	// Prepared Statements Go Here
	// example: 
	//UnPreparedStatements["exampleStmt"] = "select * from tablename where id=?"

	if !Db.IsConnected() {
		Db.Connect()
	}

	c := make(chan int)

	for stmtname, stmtsql := range UnPreparedStatements {
		go PrepareStatement(stmtname, stmtsql, c)
	}

	for _, _ = range UnPreparedStatements {
		<-c
	}

	return nil
}

func PrepareStatement(name string, sql string, ch chan int) {
	stmt, err := Db.Prepare(sql)
	if err == nil {
		Statements[name] = stmt
	}
	ch <- 1
}

func GetStatement(key string) (stmt mysql.Stmt, err error) {
	stmt, ok := Statements[key]
	if !ok {
		qry := expvar.Get(key)
		if qry == nil {
			err = errors.New("Invalid query reference")
		}
	}
	return

}
