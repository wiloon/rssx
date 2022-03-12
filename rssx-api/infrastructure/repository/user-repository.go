package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"rssx/utils/logger"
)

func test() {
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)
	createTable := `CREATE TABLE if not exists users (
  id char(36) PRIMARY KEY NOT NULL,
  name varchar(50) DEFAULT NULL,
  create_time timestamp DEFAULT NULL
);
`
	r, err := db.Exec(createTable)
	checkErr(err)
	logger.Infof("%+v", r)
	stmt, err := db.Prepare("INSERT INTO `users` VALUES (?,?,?);")
	checkErr(err)
	res, err := stmt.Exec(0, "wiloon", "2017-12-07 22:10:49")
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)
	fmt.Println(id)
	db.Close()
}

func checkErr(err error) {
	if err != nil {
		logger.Errorf("err: %v", err)
	}
}
