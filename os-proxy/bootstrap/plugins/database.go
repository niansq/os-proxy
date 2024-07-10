package plugins

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"io"
	"log"
	"os"
	"os-proxy/app/models"
	"os-proxy/bootstrap"
	"os-proxy/config"
	"os-proxy/config/plugins"
	"strconv"
	"sync"
	"time"
)

var lgDB = map[string]*LangGoDB{}

type LangGoDB struct {
	Once *sync.Once
	DB   *gorm.DB
}

func newLangGoDB() *LangGoDB {
	return &LangGoDB{
		Once: &sync.Once{},
		DB:   &gorm.DB{},
	}
}

func init() {
	p := &LangGoDB{}
	RegisterdPlugin(p)
}

// Use 切换DB
func (lg *LangGoDB) Use(dbname string) *LangGoDB {
	if db, ok := lgDB[dbname]; ok {
		return db
	} else {
		bootstrap.NewLogger().Logger.Error("切换DB失败", zap.String("当前DB名称不存在", dbname))
		panic(ok)
	}
}

func (lg *LangGoDB) NewDB() *gorm.DB {
	return lg.DB
}

func (lg *LangGoDB) Name() string {
	return "DB"
}

func (lg *LangGoDB) New() interface{} {
	conf := bootstrap.NewConfig("")
	for _, db := range conf.Database {
		lgDB[db.DBName] = newLangGoDB()
		lgDB[db.DBName].initializeDB(db, conf)
	}
	return lgDB
}

func (lg *LangGoDB) initializeDB(db *plugins.Database, conf *config.Configuration) {
	lg.Once.Do(func() {
		switch db.Driver {
		case "mysql":
			initMysqlGorm(db, conf)
		case "postgres":
			initPGGorm(db, conf)
		default:
			initMysqlGorm(db, conf)
		}
	})
}

func initPGGorm(dbConfig *plugins.Database, conf *config.Configuration) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		dbConfig.Host,
		dbConfig.UserName,
		dbConfig.Password,
		dbConfig.Database,
		strconv.Itoa(dbConfig.Port),
	)

	gormConfig := &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	if dbConfig.EnableLgLog {
		gormConfig = &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,                          // 禁用自动创建外键约束
			Logger:                                   getGormLogger(dbConfig, conf), // 使用自定义 Logger
		}
	}

	// gorm将类名转换成数据库表名的逻辑
	if gormConfig.NamingStrategy == nil {
		gormConfig.NamingStrategy = schema.NamingStrategy{
			//TablePrefix:   "t_",
			SingularTable: true,
		}
	}

	if db, err := gorm.Open(postgres.Open(dsn), gormConfig); err != nil {
		bootstrap.NewLogger().Logger.Error("mysql connect failed, err:", zap.Any("err", err))
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
		sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
		// 执行数据库脚本建表
		initMySqlTables(db)
		lgDB[dbConfig.DBName].DB = db
	}
}

func initMysqlGorm(dbConfig *plugins.Database, conf *config.Configuration) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		dbConfig.UserName,
		dbConfig.Password,
		dbConfig.Host,
		strconv.Itoa(dbConfig.Port),
		dbConfig.Database,
		dbConfig.Charset,
	)

	gormConfig := &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	if dbConfig.EnableLgLog {
		gormConfig = &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,                          // 禁用自动创建外键约束
			Logger:                                   getGormLogger(dbConfig, conf), // 使用自定义 Logger
		}
	}

	if gormConfig.NamingStrategy == nil {
		gormConfig.NamingStrategy = schema.NamingStrategy{
			//TablePrefix:   "t_",
			SingularTable: true,
		}
	}

	if db, err := gorm.Open(mysql.Open(dsn), gormConfig); err != nil {
		bootstrap.NewLogger().Logger.Error("mysql connect failed, err:", zap.Any("err", err))
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
		sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
		// 执行数据库脚本建表
		//initMySqlTables(db)
		lgDB[dbConfig.DBName].DB = db
	}
}

func getGormLogger(dbConfig *plugins.Database, conf *config.Configuration) logger.Interface {
	var logMode logger.LogLevel

	switch dbConfig.LogMode {
	case "silent":
		logMode = logger.Silent
	case "error":
		logMode = logger.Error
	case "warn":
		logMode = logger.Warn
	case "info":
		logMode = logger.Info
	default:
		logMode = logger.Info
	}

	return logger.New(getGormLogWriter(dbConfig, conf), logger.Config{
		SlowThreshold:             200 * time.Millisecond,        // 慢 SQL 阈值
		LogLevel:                  logMode,                       // 日志级别
		IgnoreRecordNotFoundError: false,                         // 忽略ErrRecordNotFound（记录未找到）错误
		Colorful:                  !dbConfig.EnableFileLogWriter, // 禁用彩色打印
	})
}

// 自定义 接管gorm日志，打印到文件 or 控制台
func getGormLogWriter(dbConfig *plugins.Database, conf *config.Configuration) logger.Writer {
	var writer io.Writer

	// 是否启用日志文件
	if dbConfig.EnableFileLogWriter {
		// 自定义 Writer
		writer = &lumberjack.Logger{
			Filename:   conf.Log.RootDir + "/" + dbConfig.LogFilename,
			MaxSize:    conf.Log.MaxSize,
			MaxBackups: conf.Log.MaxBackups,
			MaxAge:     conf.Log.MaxAge,
			Compress:   conf.Log.Compress,
		}
	} else {
		// 默认 Writer
		writer = os.Stdout
	}
	return log.New(writer, "\r\n", log.LstdFlags)
}

// 数据库表初始化,自动建表
func initMySqlTables(db *gorm.DB) {
	err := db.AutoMigrate(
		models.MetaDataInfo{},
		models.MultiPartInfo{},
		models.TaskInfo{},
		models.TaskLog{},
	)
	if err != nil {
		bootstrap.NewLogger().Logger.Error("migrate table failed", zap.Any("err", err))
		panic(err.Error())
	}
}

// 数据库一定会使用
func (lg *LangGoDB) Flag() bool {
	return true
}

func (lg *LangGoDB) Close() {}

func (lg *LangGoDB) Health() {
	for dbName, db := range lgDB {
		tx := db.DB.Exec("select now();")

		if tx.Error != nil {
			bootstrap.NewLogger().Logger.Error("db connect failed,", zap.String("当前DB名称不存在", dbName),
				zap.Any("err", tx.Error))
		}
	}
}
