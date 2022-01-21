package task

import (
	"github.com/stratosnet/sds/msg/protos"
	"github.com/stratosnet/sds/pp/file"
)

type TransferTask struct {
	FromSp           bool
	PpInfo           *protos.PPBaseInfo
	SliceStorageInfo *protos.SliceStorageInfo
}

// TransferTaskMap
var TransferTaskMap = make(map[string]TransferTask)

// CheckTransfer check whether can transfer
// todo:
func CheckTransfer(target *protos.ReqFileSliceBackupNotice) bool {
	return true
}

// GetTransferSliceData
func GetTransferSliceData(taskId string) []byte {
	if tTask, ok := TransferTaskMap[taskId]; ok {
		data := file.GetSliceData(tTask.SliceStorageInfo.SliceHash)
		return data
	}
	return nil
}

// SaveTransferData
func SaveTransferData(target *protos.RspTransferDownload) bool {
	if tTask, ok := TransferTaskMap[target.TransferCer]; ok {
		save := file.SaveSliceData(target.Data, tTask.SliceStorageInfo.SliceHash, target.Offset)
		if save {
			if target.SliceSize == uint64(file.GetSliceSize(tTask.SliceStorageInfo.SliceHash)) {
				return true
			}
			return false
		}
		return false
	}
	return false
}
