package ormengine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var (
	traceStr     = "%s\n[%.3fms] [rows:%v] %s"
	traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
	traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
)

type GormLogger struct {
	level         logger.LogLevel
	SlowThreshold time.Duration
}

func NewGormLogger(slowThreshold time.Duration) *GormLogger {
	return &GormLogger{SlowThreshold: slowThreshold}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {

	if l.level == 0 || l.level >= logger.Info {
		logx.WithContext(ctx).Infof(s, append([]interface{}{utils.FileWithLineNum()}, i...)...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	if l.level == 0 || l.level >= logger.Warn {
		logx.WithContext(ctx).Errorf(s, append([]interface{}{utils.FileWithLineNum()}, i...)...)
	}
}

func (l *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	if l.level == 0 || l.level >= logger.Error {
		logx.WithContext(ctx).Errorf(s, append([]interface{}{utils.FileWithLineNum()}, i...)...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		if rows == -1 {
			logx.WithContext(ctx).WithDuration(elapsed).Infof(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logx.WithContext(ctx).WithDuration(elapsed).Infof(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			logx.WithContext(ctx).WithDuration(elapsed).Slowf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logx.WithContext(ctx).WithDuration(elapsed).Slowf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	default:
		sql, rows := fc()
		if rows == -1 {
			logx.WithContext(ctx).WithDuration(elapsed).Infof(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			logx.WithContext(ctx).WithDuration(elapsed).Infof(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
