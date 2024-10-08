package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dbhost          = ""
	dbport          = ""
	dbuser          = ""
	dbpassword      = ""
	dbname          = ""
	accessKey       = ""
	secretAccessKey = ""
	jwtsecret       = ""
	email           = ""
	emailpassword   = ""
	resetpassword   = ""
)

// DB is connected MySQL DB
var DB *gorm.DB

func init() {
	dbhost = os.Getenv("DBHOST")
	dbport = os.Getenv("DBPORT")
	dbuser = os.Getenv("DBUSER")
	dbpassword = os.Getenv("DBPASSWORD")
	dbname = os.Getenv("DBNAME")

	accessKey = os.Getenv("ACCESSKEY")
	secretAccessKey = os.Getenv("SECRETACCESSKEY")

	jwtsecret = os.Getenv("JWTSECRET")

	email = os.Getenv("EMAIL")
	emailpassword = os.Getenv("EMAILPASSWORD")

	resetpassword = os.Getenv("RESETPASSWORD")
}

// Connect to MySQL server
func Connect() {
	fmt.Println(dbhost)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbuser,
		dbpassword,
		dbhost,
		dbport,
		dbname,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	DB = db
}

func GetCredentials() (string, string) {
	return accessKey, secretAccessKey
}

func GetJWTSecret() string {
	return jwtsecret
}

func GetEmailCredentials() (string, string) {
	return email, emailpassword
}

func GetResetPassword() string {
	return resetpassword
}
