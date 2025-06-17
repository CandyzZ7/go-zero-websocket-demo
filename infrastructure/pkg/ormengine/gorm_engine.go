package ormengine

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlConf struct {
	DataSource    string
	MaxOpenConns  int           `json:",default=10"`
	MaxIdleConns  int           `json:",default=100"`
	MaxLifetime   time.Duration `json:",default=1h"`
	SlowThreshold time.Duration `json:",default=500ms"`
}

var (
	gormEngine *gorm.DB
	once       sync.Once
)
var ErrFieldNotFound = errors.New("field not found")

func NewGormEngine(mysqlConf MysqlConf) (*gorm.DB, error) {
	var err error
	once.Do(func() {
		gormEngine, err = gorm.Open(mysql.Open(mysqlConf.DataSource), &gorm.Config{
			Logger: NewGormLogger(mysqlConf.SlowThreshold), // 调整日志级别，根据需要修改
		})
		if err != nil {
			return
		}
		db, err := gormEngine.DB()
		if err != nil {
			return
		}
		db.SetMaxIdleConns(mysqlConf.MaxIdleConns)
		db.SetMaxOpenConns(mysqlConf.MaxOpenConns)
		db.SetConnMaxLifetime(mysqlConf.MaxLifetime)
		err = gormEngine.Use(NewCustomePlugin())
		if err != nil {
			return
		}
	})
	return gormEngine, err
}

func MustNewGormEngine(mysqlConf MysqlConf) *gorm.DB {
	db, err := NewGormEngine(mysqlConf)
	if err != nil {
		logx.Must(err)
	}
	return db
}
