package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

func initGorm() {
	db = createGorm()
}

// dbs 返回一个带有新会话的 *gorm.DB 对象。
// 该函数在需要执行不影响原始数据库会话的操作时非常有用。
// 无需参数。
// 返回值是 *gorm.DB，表示一个新的数据库会话。
func dbs() *gorm.DB {
	return db.Session(&gorm.Session{})
}

func createGorm() *gorm.DB {
	var db *gorm.DB

	if Config.System.Database == "mysql" {
		loggerLevel := getLoggerLevel(Config.Mysql.LoggerLevel)
		db = mysqlConnect(loggerLevel)
	} else {
		loggerLevel := getLoggerLevel(Config.Sqlite3.LoggerLevel)
		db = sqlite3Connect(loggerLevel)
	}

	return db
}

func getLoggerLevel(level string) logger.LogLevel {
	switch level {
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	case "silent":
		return logger.Silent
	default:
		return logger.Info
	}
}

func sqlite3Connect(loggerLevel logger.LogLevel) *gorm.DB {
	// 获取数据库文件所在的目录路径
	pathDir := filepath.Dir(Config.Sqlite3.Path)

	// 检查目录是否存在，如果不存在则尝试创建
	if _, err := os.Stat(pathDir); os.IsNotExist(err) {
		// 创建多级目录，权限为 0755
		if err := os.MkdirAll(pathDir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create directory: %s, error: %s", pathDir, err.Error()))
		}
	}

	//判断数据库文件是否存在，如果存在，看是否有读写权限，没有则赋予其读写权限
	if _, err := os.Stat(Config.Sqlite3.Path); !os.IsNotExist(err) {
		if err := os.Chmod(Config.Sqlite3.Path, 0664); err != nil {
			fmt.Println("Failed to set file permissions:", err.Error())
		}
	}

	db, err := gorm.Open(sqlite.Open(Config.Sqlite3.Path), &gorm.Config{
		//SkipDefaultTransaction: true,                                //跳过默认事务
		PrepareStmt:                              true,                                //预编译语句
		Logger:                                   logger.Default.LogMode(loggerLevel), //打印全部sql日志
		DisableForeignKeyConstraintWhenMigrating: true,                                // 禁用外键(用于禁用默认外键设置)
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,  //使用单数表名
			NoLowerCase:   false, // 关闭小写转换
		},
	})
	if err != nil {
		panic("failed to connect database")
	}

	db.Exec("PRAGMA journal_mode=WAL;")
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Minute * 10)

	return db
}

func mysqlConnect(loggerLevel logger.LogLevel) *gorm.DB {
	db_dns := Config.Mysql.Dsn()
	db, err := gorm.Open(mysql.Open(db_dns), &gorm.Config{
		//SkipDefaultTransaction: true, //跳过默认事务
		PrepareStmt:                              true,                                //预编译语句
		Logger:                                   logger.Default.LogMode(loggerLevel), //打印全部sql日志
		DisableForeignKeyConstraintWhenMigrating: true,                                // 禁用外键(用于禁用默认外键设置)
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,  //使用单数表名
			NoLowerCase:   false, // 关闭小写转换
		},
	})
	if err != nil {
		panic("mysql 链接失败：" + db_dns)
	}
	sqlDb, dbErr := db.DB()
	if dbErr != nil {
		panic("mysql DB 链接失败：" + dbErr.Error())
	}
	sqlDb.SetMaxIdleConns(Config.Mysql.MaxIdleConns)
	sqlDb.SetMaxOpenConns(Config.Mysql.MaxOpenConns)
	return db

}
