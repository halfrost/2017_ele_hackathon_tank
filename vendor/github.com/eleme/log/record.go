package log

import (
	"runtime"
	"strconv"
	"time"
)

// Record represents a record object.
type Record interface {
	Level() LevelType
	AppID() string
	Now() time.Time
	Name() string
	Fileline() string
}

// RecordFactory represents a factory of record.
type RecordFactory func(name string, calldepth int, lv LevelType, msg string) Record

// BaseRecord stands for a single record of log, usually a single line
type BaseRecord struct {
	fileLine string
	name     string
	now      time.Time
	lv       LevelType
	msg      string

	appID string
}

// Level return the level of BaseRecord.
func (br *BaseRecord) Level() LevelType {
	return br.lv
}

// AppID return the appID of BaseRecord.
func (br *BaseRecord) AppID() string {
	return br.appID
}

// Now return the now of time of BaseRecord.
func (br *BaseRecord) Now() time.Time {
	return br.now
}

// Name return the name of BaseRecord.
func (br *BaseRecord) Name() string {
	return br.name
}

// Fileline return the name of BaseRecord.
func (br *BaseRecord) Fileline() string {
	return br.fileLine
}

// String returns the raw message of the Record
func (br *BaseRecord) String() string {
	return br.msg
}

// NewBaseRecord creates a BaseRecord with given name, calldepth, LevelType and msg.
func NewBaseRecord(name string, calldepth int, lv LevelType, msg string) *BaseRecord {
	fileLine := ""
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	fileLine = file + ":" + strconv.Itoa(line)

	r := &BaseRecord{
		fileLine: fileLine,
		name:     name,
		now:      time.Now(),
		lv:       lv,
		msg:      msg,
		appID:    globalAppID,
	}
	return r
}

// NewBaseRecordFactory return a record factory for BaseRecord.
func NewBaseRecordFactory() RecordFactory {
	return func(name string, calldepth int, lv LevelType, msg string) Record {
		return NewBaseRecord(name, calldepth+1, lv, msg)
	}
}
