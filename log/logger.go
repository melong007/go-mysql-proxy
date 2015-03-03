// Copyright (c) 2014 Xiaomi.com, Inc. All Rights Reserved
// @file    logger.go
// @author  王靖 (wangjing1@xiaomi.com)
// @date    14-11-25 20:02:50
// @version $Revision: 1.0 $
// @brief

package log

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"micode.be.xiaomi.com/golib/milog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// SysLog 系统Log
var SysLog *Logger = nil

// AppLog 应用Log
var AppLog *Logger = nil

// Logger the milog.Logger wrapper
type Logger struct {
	l *milog.Logger
}

func logidGenerator() string {
	if i, err := rand.Int(rand.Reader, big.NewInt(1<<30-1)); err != nil {
		return "0"
	} else {
		return i.String()
	}
}

func comMessage(strfmt string, args ...interface{}) map[string]string {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "?"
		line = 0
	}
	fn := runtime.FuncForPC(pc)
	var fnName string
	if fn == nil {
		fnName = "?()"
	} else {
		dotName := filepath.Ext(fn.Name())
		fnName = strings.TrimLeft(dotName, ".") + "()"
	}
	ret := map[string]string{
		"file": filepath.Base(file) + ":" + strconv.Itoa(line),
		"func": fnName,
		"msg":  fmt.Sprintf(strfmt, args...),
	}

	return ret
}

// Notice print notice message to logfile
func (lg *Logger) Notice(strfmt string, args ...interface{}) {
	lg.l.Notice(comMessage(strfmt, args...), logidGenerator())
}

// Debug print debug message to logfile
func (lg *Logger) Debug(strfmt string, args ...interface{}) {
	lg.l.Debug(comMessage(strfmt, args...), logidGenerator())
}

// Warn print warning message to logfile
func (lg *Logger) Warn(strfmt string, args ...interface{}) {
	lg.l.Warn(comMessage(strfmt, args...), logidGenerator())
}

// Fatal print fatal message to logfile
func (lg *Logger) Fatal(strfmt string, args ...interface{}) {
	lg.l.Fatal(comMessage(strfmt, args...), logidGenerator())
}

// Config Config of One Log Instance
type Config struct {
	FilePath string
	LogLevel int
	AppTag   string
}

func init() {
	realInit(&Config{FilePath: "/dev/stdout", LogLevel: 0},
		&Config{FilePath: "/dev/stdout", LogLevel: 0})
}

var once sync.Once

func Init(syslog, applog *Config) {
	f := func() {
		realInit(syslog, applog)
	}
	once.Do(f)
}

func realInit(syslog, applog *Config) {
	SysLog = &Logger{
		l: milog.NewLogger(syslog.FilePath),
	}
	SysLog.l.SetLevel(syslog.LogLevel)
	SysLog.l.SetAppTag(defaultAppTag())

	AppLog = &Logger{
		l: milog.NewLogger(applog.FilePath),
	}
	AppLog.l.SetLevel(applog.LogLevel)
	AppLog.l.SetAppTag(defaultAppTag())
}

func defaultAppTag() string {
	return "mysql-proxy"
}

/* vim: set expandtab ts=4 sw=4 */
