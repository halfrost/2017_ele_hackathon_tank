package rpc

import (
	"github.com/eleme/log"
)

// ELogRecord represents a ELogRecord with rpcId and requestID.
type ELogRecord struct {
	*log.BaseRecord
	rpcID     string
	requestID string
}

// NewELogRecord create a ELogRecord with rpcID and requestID.
func NewELogRecord(name string, calldepth int, lv log.LevelType, msg string, rpcID string, requestID string) *ELogRecord {
	return &ELogRecord{
		BaseRecord: log.NewBaseRecord(name, calldepth, lv, msg),
		rpcID:      rpcID,
		requestID:  requestID,
	}
}

// NewELogRecordFactory return a record factory for RPCRecord.
func NewELogRecordFactory(rpcID string, requestID string) log.RecordFactory {
	return func(name string, calldepth int, lv log.LevelType, msg string) log.Record {
		return NewELogRecord(name, calldepth+2, lv, msg, rpcID, requestID)
	}
}
