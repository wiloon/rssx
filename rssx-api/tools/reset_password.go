package main

import (
	"flag"
	"fmt"
	"log"

	"rssx/utils/config"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	Id         string
	Name       string
	Password   string
	CreateTime string
}

func main() {
	username := flag.String("user", "", "用户名")
	password := flag.String("password", "", "新密码")
	dbPath := flag.String("db", "", "数据库路径（默认使用配置文件中的路径）")
	genHash := flag.Bool("gen", false, "只生成密码哈希，不更新数据库")

	flag.Parse()

	if *password == "" {
		log.Fatal("请指定密码: -password <新密码>")
	}

	// 生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("生成密码哈希失败: %v", err)
	}

	fmt.Printf("密码: %s\n", *password)
	fmt.Printf("哈希: %s\n", string(hash))

	// 如果只是生成哈希，到这里就结束
	if *genHash {
		return
	}

	// 需要用户名才能更新数据库
	if *username == "" {
		log.Fatal("请指定用户名: -user <用户名>")
	}

	// 获取数据库路径：优先使用命令行参数，否则从配置文件读取
	finalDbPath := *dbPath
	if finalDbPath == "" {
		finalDbPath = config.GetString("sqlite.path", "./data/rssx-api.db")
	}
	fmt.Printf("数据库路径: %s\n", finalDbPath)

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(finalDbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 更新密码
	result := db.Model(&User{}).Where("name = ?", *username).Update("password", string(hash))
	if result.Error != nil {
		log.Fatalf("更新密码失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Printf("警告: 用户 '%s' 不存在", *username)
	} else {
		fmt.Printf("成功更新用户 '%s' 的密码\n", *username)
	}
}
