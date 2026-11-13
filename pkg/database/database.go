package database

import (
	"fmt"
	"log"
	"os"
	"time"

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
}

// Connect to MySQL server
func Connect() {
	// interpolateParams sends a query with its args in one round trip instead of prepare, execute and close
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true",
		dbuser,
		dbpassword,
		dbhost,
		dbport,
		dbname,
	)
	// Multi statement writes use explicit transactions, single statements don't need one
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		log.Fatal(err)
	}

	// Lambda instance handles one request at a time, keep the pool small and drop stale connections
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Minute)

	DB = db
}

func GetCredentials() (string, string) {
	return accessKey, secretAccessKey
}

func GetJWTSecret() string {
	return jwtsecret
}
