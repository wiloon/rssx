package data

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "user0:password0@tcp(192.168.1.108:3306)/rssx?charset=utf8")
	if err != nil {
		log.Println("failed to connect to db;", err)
	}
}

func Find(feedId int) {
	rows, err := db.Query("SELECT title FROM news where feed_id=0")
	checkErr(err)

	for rows.Next() {
		var title string

		err = rows.Scan(&title)
		checkErr(err)

		log.Println(title)
	}

}
func Save(title, url, description string) {

	//插入数据
	stmt, err := db.Prepare("INSERT news SET feed_id=?,title=?,url=?,description=?")
	checkErr(err)

	res, err := stmt.Exec("0", title, url, description)
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	log.Println(id)
	////更新数据
	//stmt, err = db.Prepare("update u set username=? where uid=?")
	//checkErr(err)
	//
	//res, err = stmt.Exec("astaxieupdate", id)
	//checkErr(err)
	//
	//affect, err := res.RowsAffected()
	//checkErr(err)
	//
	//fmt.Println(affect)

	////查询数据
	//rows, err := db.Query("SELECT * FROM userinfo")
	//checkErr(err)
	//
	//for rows.Next() {
	//	var uid int
	//	var username string
	//	var department string
	//	var created string
	//	err = rows.Scan(&uid, &username, &department, &created)
	//	checkErr(err)
	//	fmt.Println(uid)
	//	fmt.Println(username)
	//	fmt.Println(department)
	//	fmt.Println(created)
	//}
	//
	////删除数据
	//stmt, err = db.Prepare("delete from userinfo where uid=?")
	//checkErr(err)
	//
	//res, err = stmt.Exec(id)
	//checkErr(err)
	//
	//affect, err = res.RowsAffected()
	//checkErr(err)
	//
	//fmt.Println(affect)

}

func Close() {
	db.Close()
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
