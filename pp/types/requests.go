package types

import (
	"context"
	"crypto/ed25519"
	"math"
	"path"
	"reflect"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/stratosnet/sds/framework/core"
	"github.com/stratosnet/sds/msg"
	"github.com/stratosnet/sds/msg/header"
	"github.com/stratosnet/sds/msg/protos"
	"github.com/stratosnet/sds/pp/file"
	"github.com/stratosnet/sds/pp/setting"
	"github.com/stratosnet/sds/pp/task"
	"github.com/stratosnet/sds/utils"
)

func ReqRegisterData() *protos.ReqRegister {
	return &protos.ReqRegister{
		Address: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
		MyAddress: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
		PublicKey: setting.P2PPublicKey,
	}
}

func ReqRegisterDataTR(target *protos.ReqRegister) *msg.RelayMsgBuf {
	req := target
	req.MyAddress = &protos.PPBaseInfo{
		P2PAddress:     setting.P2PAddress,
		WalletAddress:  setting.WalletAddress,
		NetworkAddress: setting.NetworkAddress,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		utils.ErrorLog(err)
	}
	return &msg.RelayMsgBuf{
		MSGHead: PPMsgHeader(data, header.ReqRegister),
		MSGData: data,
	}
}

func ReqMiningData() *protos.ReqMining {
	return &protos.ReqMining{
		Address: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
			RestAddress:    setting.RestAddress,
		},
		PublicKey: setting.P2PPublicKey,
		Sign:      setting.GetSign(setting.P2PAddress),
	}
}

func ReqGetPPlistData() *protos.ReqGetPPList {
	return &protos.ReqGetPPList{
		MyAddress: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
	}
}

func ReqGetSPlistData() *protos.ReqGetSPList {
	return &protos.ReqGetSPList{
		MyAddress: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
	}
}

// RequestUploadFileData RequestUploadFileData
func RequestUploadFileData(paths, storagePath, reqID string, isCover bool, isVideoStream bool) *protos.ReqUploadFile {
	info := file.GetFileInfo(paths)
	if info == nil {
		utils.ErrorLog("wrong filePath")
		return nil
	}
	fileName := info.Name()
	utils.Log("fileName~~~~~~~~~~~~~~~~~~~~~~~~", fileName)
	fileHash := file.GetFileHash(paths)
	utils.Log("fileHash~~~~~~~~~~~~~~~~~~~~~~", fileHash)

	p2pFileString := setting.P2PAddress + fileHash

	req := &protos.ReqUploadFile{
		FileInfo: &protos.FileInfo{
			FileSize:    uint64(info.Size()),
			FileName:    fileName,
			FileHash:    fileHash,
			StoragePath: storagePath,
		},
		MyAddress: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
		Sign:          setting.GetSign(p2pFileString),
		IsCover:       isCover,
		ReqId:         reqID,
		IsVideoStream: isVideoStream,
	}
	if isCover {
		fileSuffix := path.Ext(paths)
		req.FileInfo.FileName = fileHash + fileSuffix
	}
	if isVideoStream {
		duration, err := file.GetVideoDuration(paths)
		if err != nil {
			utils.ErrorLog("Failed to get the length of the video: ", err)
			return nil
		}
		req.FileInfo.Duration = duration
	}
	p2pFileHash := []byte(p2pFileString)
	utils.DebugLogf("setting.WalletAddress + fileHash : %v", p2pFileHash)

	if ed25519.Verify(setting.P2PPublicKey, p2pFileHash, req.Sign) {
		utils.DebugLog("ECC verification ok")
	} else {
		utils.DebugLog("ECC verification failed")
	}

	// info
	p := &task.UpProgress{
		Total:     info.Size(),
		HasUpload: 0,
	}
	task.UpLoadProgressMap.Store(fileHash, p)
	// if isCover {
	// 	os.Remove(path)
	// }
	return req
}

func RspDownloadSliceData(target *protos.ReqDownloadSlice) *protos.RspDownloadSlice {
	slice := task.GetDownloadSlice(target)
	return &protos.RspDownloadSlice{
		P2PAddress:    target.P2PAddress,
		WalletAddress: target.WalletAddress,
		SliceInfo:     target.SliceInfo,
		FileCrc:       slice.FileCrc,
		FileHash:      target.FileHash,
		TaskId:        target.TaskId,
		Data:          slice.Data,
		SliceSize:     uint64(len(slice.Data)),
		SavePath:      target.SavePath,
		ReqId:         target.ReqId,
	}
}

func RspDownloadSliceDataSplit(rsp *protos.RspDownloadSlice, dataStart, dataEnd, offsetStart, offsetEnd uint64, last bool) *protos.RspDownloadSlice {
	if dataEnd == 0 {
		return &protos.RspDownloadSlice{
			SliceInfo: &protos.SliceOffsetInfo{
				SliceHash: rsp.SliceInfo.SliceHash,
				SliceOffset: &protos.SliceOffset{
					SliceOffsetStart: offsetStart,
					SliceOffsetEnd:   offsetEnd,
				},
			},
			FileCrc:       rsp.FileCrc,
			FileHash:      rsp.FileHash,
			Data:          rsp.Data[dataStart:],
			P2PAddress:    rsp.P2PAddress,
			WalletAddress: rsp.WalletAddress,
			TaskId:        rsp.TaskId,
			SliceSize:     rsp.SliceSize,
			Result:        rsp.Result,
			NeedReport:    last,
			SavePath:      rsp.SavePath,
			ReqId:         rsp.ReqId,
		}
	}
	return &protos.RspDownloadSlice{
		SliceInfo: &protos.SliceOffsetInfo{
			SliceHash: rsp.SliceInfo.SliceHash,
			SliceOffset: &protos.SliceOffset{
				SliceOffsetStart: offsetStart,
				SliceOffsetEnd:   offsetEnd,
			},
		},
		FileCrc:       rsp.FileCrc,
		FileHash:      rsp.FileHash,
		Data:          rsp.Data[dataStart:dataEnd],
		P2PAddress:    rsp.P2PAddress,
		WalletAddress: rsp.WalletAddress,
		TaskId:        rsp.TaskId,
		SliceSize:     rsp.SliceSize,
		Result:        rsp.Result,
		NeedReport:    last,
		SavePath:      rsp.SavePath,
		ReqId:         rsp.ReqId,
	}

}

func ReqUploadFileSliceData(task *task.UploadSliceTask) *protos.ReqUploadFileSlice {
	return &protos.ReqUploadFileSlice{
		TaskId:        task.TaskID,
		FileCrc:       task.FileCRC,
		SliceNumAddr:  task.SliceNumAddr,
		SliceInfo:     task.SliceOffsetInfo,
		Data:          task.Data,
		FileHash:      task.FileHash,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		SliceSize:     task.SliceTotalSize,
	}
}

func ReqReportUploadSliceResultData(target *protos.RspUploadFileSlice) *protos.ReportUploadSliceResult {

	utils.DebugLog("reqReportUploadSliceResultData____________________", target.SliceSize)
	return &protos.ReportUploadSliceResult{
		TaskId:        target.TaskId,
		SliceNumAddr:  target.SliceNumAddr,
		SliceHash:     target.SliceHash,
		IsPP:          false,
		UploadSuccess: true,
		FileHash:      target.FileHash,
		SliceSize:     target.SliceSize,
		Sign:          setting.GetSign(setting.P2PAddress + target.FileHash),
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
	}
}
func ReqReportUploadSliceResultDataPP(target *protos.ReqUploadFileSlice) *protos.ReportUploadSliceResult {
	utils.DebugLog("____________________", target.SliceSize)
	return &protos.ReportUploadSliceResult{
		TaskId:        target.TaskId,
		SliceNumAddr:  target.SliceNumAddr,
		SliceHash:     target.SliceInfo.SliceHash,
		IsPP:          true,
		UploadSuccess: true,
		FileHash:      target.FileHash,
		SliceSize:     target.SliceSize,
		Sign:          setting.GetSign(setting.P2PAddress + target.FileHash),
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
	}
}

func RspUploadFileSliceData(target *protos.ReqUploadFileSlice) *protos.RspUploadFileSlice {
	return &protos.RspUploadFileSlice{
		TaskId:        target.TaskId,
		FileHash:      target.FileHash,
		SliceHash:     target.SliceInfo.SliceHash,
		P2PAddress:    target.P2PAddress,
		WalletAddress: target.WalletAddress,
		SliceNumAddr:  target.SliceNumAddr,
		SliceSize:     target.SliceSize,
		Result: &protos.Result{
			State: protos.ResultState_RES_SUCCESS,
		},
	}
}

func ReqReportDownloadResultData(target *protos.RspDownloadSlice, isPP bool) *protos.ReqReportDownloadResult {

	utils.DebugLog("#################################################################", target.SliceInfo.SliceHash)
	repReq := &protos.ReqReportDownloadResult{
		IsPP:                    isPP,
		DownloaderP2PAddress:    target.P2PAddress,
		DownloaderWalletAddress: target.WalletAddress,
		MyP2PAddress:            setting.P2PAddress,
		MyWalletAddress:         setting.WalletAddress,
		FileHash:                target.FileHash,
		Sign:                    setting.GetSign(setting.P2PAddress + target.FileHash),
		TaskId:                  target.TaskId,
	}
	if isPP {
		utils.Log("PP ReportDownloadResult ")
		if dlTask, ok := task.DownloadTaskMap.Load(target.FileHash + target.WalletAddress); ok {
			downloadTask := dlTask.(*task.DownloadTask)
			utils.DebugLog("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^downloadTask", downloadTask)
			if sInfo, ok := downloadTask.SliceInfo[target.SliceInfo.SliceHash]; ok {
				repReq.SliceInfo = sInfo
				repReq.SliceInfo.VisitResult = true
			} else {
				utils.DebugLog("ReportDownloadResult failed~~~~~~~~~~~~~~~~~~~~~~~~~~")
			}

		} else {
			repReq.SliceInfo = &protos.DownloadSliceInfo{
				SliceStorageInfo: &protos.SliceStorageInfo{
					SliceHash: target.SliceInfo.SliceHash,
				},
			}
		}
	} else {
		repReq.SliceInfo = &protos.DownloadSliceInfo{
			SliceStorageInfo: &protos.SliceStorageInfo{
				SliceHash: target.SliceInfo.SliceHash,
			},
		}
	}
	return repReq
}

func ReqDownloadSliceData(target *protos.RspFileStorageInfo, rsp *protos.DownloadSliceInfo) *protos.ReqDownloadSlice {
	return &protos.ReqDownloadSlice{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		FileHash:      target.FileHash,
		TaskId:        rsp.TaskId,
		SliceInfo: &protos.SliceOffsetInfo{
			SliceHash:   rsp.SliceStorageInfo.SliceHash,
			SliceOffset: rsp.SliceOffset,
		},
		SavePath: target.SavePath,
		ReqId:    uuid.New().String(),
	}
}

func RspFileStorageInfoData(target *protos.RspFileStorageInfo) *msg.RelayMsgBuf {

	utils.DebugLog("download detail，", target)
	sendTarget := target
	sliceInfoArr := []*protos.DownloadSliceInfo{}
	for _, info := range sendTarget.SliceInfo {
		newInfo := protos.DownloadSliceInfo{
			SliceStorageInfo: info.SliceStorageInfo,
			SliceNumber:      info.SliceNumber,
			VisitResult:      info.VisitResult,
			TaskId:           info.TaskId,
			SliceOffset:      info.SliceOffset,
		}
		sliceInfoArr = append(sliceInfoArr, &newInfo)
	}
	sendTarget.SliceInfo = sliceInfoArr
	sendTarget.RestAddress = setting.RestAddress
	sendData, err := proto.Marshal(sendTarget)
	if err != nil {
		utils.ErrorLog(err)
	}
	return &msg.RelayMsgBuf{
		MSGData: sendData,
		MSGHead: PPMsgHeader(sendData, header.RspFileStorageInfo),
	}
}

func ReqRegisterNewPPData() *protos.ReqRegisterNewPP {
	sysInfo := utils.GetSysInfo()
	return &protos.ReqRegisterNewPP{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		DiskSize:      sysInfo.DiskSize,
		MemorySize:    sysInfo.MemorySize,
		OsAndVer:      sysInfo.OSInfo,
		CpuInfo:       sysInfo.CPUInfo,
		MacAddress:    sysInfo.MacAddress,
		Version:       setting.Config.Version,
		PubKey:        setting.P2PPublicKey,
		Sign:          setting.GetSign(setting.P2PAddress),
	}
}

func ReqValidateTransferCerData(target *protos.ReqTransferNotice) *protos.ReqValidateTransferCer {
	return &protos.ReqValidateTransferCer{
		TransferCer: target.TransferCer,
		NewPp:       target.StoragePpInfo,
		OriginalPp: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
	}
}

func ReqTransferNoticeData(target *protos.ReqTransferNotice) *msg.RelayMsgBuf {
	sendTager := &protos.ReqTransferNotice{
		FromSp:      false,
		TransferCer: target.TransferCer,
		StoragePpInfo: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},

		SliceStorageInfo: target.SliceStorageInfo,
		DeleteOrigin:     target.DeleteOrigin,
	}
	data, err := proto.Marshal(sendTager)
	if err != nil {
		utils.ErrorLog(err)
	}
	return &msg.RelayMsgBuf{
		MSGHead: PPMsgHeader(data, header.ReqTransferNotice),
		MSGData: data,
	}
}

func RspTransferNoticeData(agree bool, cer string) *protos.RspTransferNotice {
	rsp := &protos.RspTransferNotice{
		StoragePpInfo: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
		TransferCer: cer,
	}
	if agree {
		rsp.Result = &protos.Result{
			State: protos.ResultState_RES_SUCCESS,
		}
	} else {
		rsp.Result = &protos.Result{
			State: protos.ResultState_RES_FAIL,
		}
	}
	return rsp
}

func ReqTransferDownloadData(transferCer string) *protos.ReqTransferDownload {
	return &protos.ReqTransferDownload{
		TransferCer: transferCer,
	}
}

//TODO: Change to BP to SP
func ReqReportTaskBPData(taskID string, traffic uint64) *msg.RelayMsgBuf {
	utils.DebugLog("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~reqReportTaskBPDatareqReportTaskBPData  taskID ==", taskID, "traffic == ", traffic)
	sendTager := &protos.ReqReportTaskBP{
		TaskId:  taskID,
		Traffic: traffic,
		Reporter: &protos.PPBaseInfo{
			P2PAddress:     setting.P2PAddress,
			WalletAddress:  setting.WalletAddress,
			NetworkAddress: setting.NetworkAddress,
		},
	}
	data, err := proto.Marshal(sendTager)
	if err != nil {
		utils.ErrorLog(err)
	}
	return &msg.RelayMsgBuf{
		MSGHead: PPMsgHeader(data, header.ReqReportTaskBP),
		MSGData: data,
	}
}

func ReqFileStorageInfoData(path, savePath, reqID string, isVideoStream bool, shareRequest *protos.ReqGetShareFile) *protos.ReqFileStorageInfo {
	return &protos.ReqFileStorageInfo{
		FileIndexes: &protos.FileIndexes{
			P2PAddress:    setting.P2PAddress,
			WalletAddress: setting.WalletAddress,
			FilePath:      path,
			SavePath:      savePath,
		},
		Sign:          setting.GetSign(setting.P2PAddress + path),
		ReqId:         reqID,
		IsVideoStream: isVideoStream,
		ShareRequest:  shareRequest,
	}
}

func FindMyFileListData(fileName, dir, reqID, keyword string, fileType protos.FileSortType, isUp bool) *protos.ReqFindMyFileList {
	return &protos.ReqFindMyFileList{
		FileName:      fileName,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		Directory:     dir,
		ReqId:         reqID,
		FileType:      fileType,
		IsUp:          isUp,
		Keyword:       keyword,
	}
}

func FindDirectoryData(reqID string) *protos.ReqFindDirectory {
	return &protos.ReqFindDirectory{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
	}
}

func FileSortData(files []*protos.FileInfo, reqID, albumID string) *protos.ReqFileSort {
	return &protos.ReqFileSort{
		Files:         files,
		ReqId:         reqID,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		AlbumId:       albumID,
	}
}

func RspTransferDownloadResultData(transferCer string) *protos.RspTransferDownloadResult {
	return &protos.RspTransferDownloadResult{
		TransferCer: transferCer,
		Result: &protos.Result{
			State: protos.ResultState_RES_SUCCESS,
		},
	}
}

func RspTransferDownload(data []byte, transferCer string, offset, sliceSize uint64) *protos.RspTransferDownload {
	return &protos.RspTransferDownload{
		Data:        data,
		TransferCer: transferCer,
		Offset:      offset,
		SliceSize:   sliceSize,
	}
}

func ReqDeleteFileData(fileHash, reqID string) *protos.ReqDeleteFile {
	return &protos.ReqDeleteFile{
		FileHash:      fileHash,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		Sign:          setting.GetSign(setting.P2PAddress + fileHash),
		ReqId:         reqID,
	}
}

func ReqDownloadSliceWrong(taskID, sliceHash, p2pAddress, walletAddress string, wrongType protos.DownloadWrongType) *protos.ReqDownloadSliceWrong {
	return &protos.ReqDownloadSliceWrong{
		P2PAddress:    p2pAddress,
		WalletAddress: walletAddress,
		TaskId:        taskID,
		SliceHash:     sliceHash,
		Type:          wrongType,
	}
}

func RspDownloadSliceWrong(target *protos.RspDownloadSliceWrong) *msg.RelayMsgBuf {
	sendTager := &protos.ReqDownloadSlice{
		SliceInfo: &protos.SliceOffsetInfo{
			SliceHash:   target.NewSliceInfo.SliceStorageInfo.SliceHash,
			SliceOffset: target.NewSliceInfo.SliceOffset,
		},
		P2PAddress:    target.P2PAddress,
		WalletAddress: target.WalletAddress,
		TaskId:        target.TaskId,
		FileHash:      target.FileHash,
	}
	data, err := proto.Marshal(sendTager)
	if err != nil {
		utils.ErrorLog(err)
	}
	return &msg.RelayMsgBuf{
		MSGHead: PPMsgHeader(data, header.ReqDownloadSlice),
		MSGData: data,
	}
}

func RspGetHDInfoData(total, free uint64) *protos.RspGetHDInfo {

	return &protos.RspGetHDInfo{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		DiskSize:      total,
		DiskFree:      free,
	}
}

func RspDeleteSliceData(sliceHash, msg string, result bool) *protos.RspDeleteSlice {
	state := protos.ResultState_RES_SUCCESS
	if !result {
		state = protos.ResultState_RES_FAIL
	}
	return &protos.RspDeleteSlice{
		P2PAddress: setting.P2PAddress,
		SliceHash:  sliceHash,
		Result: &protos.Result{
			State: state,
			Msg:   msg,
		},
	}
}

func ReqMakeDirectoryData(path, reqID string) *protos.ReqMakeDirectory {
	return &protos.ReqMakeDirectory{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		Directory:     path,
		ReqId:         reqID,
	}
}

func ReqRemoveDirectoryData(path, reqID string) *protos.ReqRemoveDirectory {
	return &protos.ReqRemoveDirectory{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		Directory:     path,
		ReqId:         reqID,
	}
}

func ReqShareLinkData(reqID string) *protos.ReqShareLink {
	return &protos.ReqShareLink{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
	}
}

func ReqShareFileData(reqID, fileHash, pathHash string, isPrivate bool, shareTime int64) *protos.ReqShareFile {
	return &protos.ReqShareFile{
		FileHash:      fileHash,
		IsPrivate:     isPrivate,
		ShareTime:     shareTime,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		PathHash:      pathHash,
		ReqId:         reqID,
	}
}

func ReqDeleteShareData(reqID, shareID string) *protos.ReqDeleteShare {
	return &protos.ReqDeleteShare{
		ReqId:         reqID,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ShareId:       shareID,
	}
}

func ReqSaveFileData(fileHash, reqID, ownerAddress string) *protos.ReqSaveFile {
	return &protos.ReqSaveFile{
		FileHash:               fileHash,
		FileOwnerWalletAddress: ownerAddress,
		P2PAddress:             setting.P2PAddress,
		WalletAddress:          setting.WalletAddress,
		ReqId:                  reqID,
	}

}

func ReqSaveFolderData(folderHash, reqID, ownerAddress string) *protos.ReqSaveFolder {
	return &protos.ReqSaveFolder{
		FolderHash:               folderHash,
		FolderOwnerWalletAddress: ownerAddress,
		P2PAddress:               setting.P2PAddress,
		WalletAddress:            setting.WalletAddress,
		ReqId:                    reqID,
	}

}

func ReqMoveFileDirectoryData(fileHash, originalDir, targetDir, reqID string) *protos.ReqMoveFileDirectory {
	return &protos.ReqMoveFileDirectory{
		FileHash:          fileHash,
		P2PAddress:        setting.P2PAddress,
		WalletAddress:     setting.WalletAddress,
		ReqId:             reqID,
		DirectoryTarget:   targetDir,
		DirectoryOriginal: originalDir,
	}
}

func ReqGetMyConfig(p2pAddress, walletAddress, reqID string) *protos.ReqConfig {
	return &protos.ReqConfig{
		P2PAddress:     p2pAddress,
		WalletAddress:  walletAddress,
		ReqId:          reqID,
		NetworkAddress: setting.NetworkAddress,
	}
}

func ReqDownloadSlicePause(fileHash, reqID string) *protos.ReqDownloadSlicePause {
	return &protos.ReqDownloadSlicePause{
		FileHash:      fileHash,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
	}
}

func RspDownloadSlicePauseData(target *protos.ReqDownloadSlicePause) *msg.RelayMsgBuf {
	sendTager := &protos.RspDownloadSlicePause{
		P2PAddress:    target.P2PAddress,
		WalletAddress: target.WalletAddress,
		FileHash:      target.FileHash,
		ReqId:         target.ReqId,
		Result: &protos.Result{
			State: protos.ResultState_RES_SUCCESS,
		},
	}
	data, err := proto.Marshal(sendTager)
	if err != nil {
		utils.ErrorLog(err)
	}
	return &msg.RelayMsgBuf{
		MSGHead: PPMsgHeader(data, header.RspDownloadSlicePause),
		MSGData: data,
	}
}

func ReqCreateAlbumData(albumName, albumBlurb, albumCoverHash, reqID string, albumType protos.AlbumType, files []*protos.FileInfo, isPrivate bool) *protos.ReqCreateAlbum {
	return &protos.ReqCreateAlbum{
		P2PAddress:     setting.P2PAddress,
		WalletAddress:  setting.WalletAddress,
		ReqId:          reqID,
		AlbumName:      albumName,
		AlbumBlurb:     albumBlurb,
		AlbumCoverHash: albumCoverHash,
		AlbumType:      albumType,
		FileInfo:       files,
		IsPrivate:      isPrivate,
	}
}

func ReqGetShareFileData(keyword, sharePassword, reqID string) *protos.ReqGetShareFile {
	return &protos.ReqGetShareFile{
		Keyword:       keyword,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		SharePassword: sharePassword,
	}
}

func ReqFindMyAlbumData(albumType protos.AlbumType, reqID string, page, number uint64, keyword string) *protos.ReqFindMyAlbum {
	return &protos.ReqFindMyAlbum{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		AlbumType:     albumType,
		Page:          page,
		Number:        number,
		Keyword:       keyword,
	}
}

func ReqEditAlbumData(albumID, albumCoverHash, albumName, albumBlurb, reqID string, changeFiles []*protos.FileInfo, isPrivate bool) *protos.ReqEditAlbum {
	return &protos.ReqEditAlbum{
		P2PAddress:     setting.P2PAddress,
		WalletAddress:  setting.WalletAddress,
		ReqId:          reqID,
		AlbumId:        albumID,
		AlbumCoverHash: albumCoverHash,
		AlbumName:      albumName,
		AlbumBlurb:     albumBlurb,
		ChangeFiles:    changeFiles,
		IsPrivate:      isPrivate,
	}
}

func ReqAlbumContentData(albumID, reqID string) *protos.ReqAlbumContent {
	return &protos.ReqAlbumContent{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		AlbumId:       albumID,
	}
}

func ReqCollectionAlbumData(albumID, reqID string, isCollection bool) *protos.ReqCollectionAlbum {
	return &protos.ReqCollectionAlbum{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		AlbumId:       albumID,
		IsCollection:  isCollection,
	}
}

func ReqDeleteAlbumData(albumID, reqID string) *protos.ReqDeleteAlbum {
	return &protos.ReqDeleteAlbum{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		AlbumId:       albumID,
	}
}
func ReqSearchAlbumData(keyword, reqID string, aType protos.AlbumType, sType protos.AlbumSortType, page, number uint64) *protos.ReqSearchAlbum {
	return &protos.ReqSearchAlbum{
		AlbumType:     aType,
		Keyword:       keyword,
		AlbumSortType: sType,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		Page:          page,
		Number:        number,
	}
}

func ReqInviteData(code, reqID string) *protos.ReqInvite {
	return &protos.ReqInvite{
		P2PAddress:     setting.P2PAddress,
		WalletAddress:  setting.WalletAddress,
		ReqId:          reqID,
		InvitationCode: code,
	}
}
func ReqGetRewardData(reqID string) *protos.ReqGetReward {
	return &protos.ReqGetReward{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
	}
}

func ReqAbstractAlbumData(reqID string) *protos.ReqAbstractAlbum {
	return &protos.ReqAbstractAlbum{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
	}
}

func ReqMyCollectionAlbumData(aType protos.AlbumType, reqID string, page, number uint64, keyword string) *protos.ReqMyCollectionAlbum {
	return &protos.ReqMyCollectionAlbum{
		AlbumType:     aType,
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		Page:          page,
		Number:        number,
		Keyword:       keyword,
	}

}

func ReqFindDirectoryTreeData(reqID, pathHash string) *protos.ReqFindDirectoryTree {
	return &protos.ReqFindDirectoryTree{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
		PathHash:      pathHash,
	}

}

func UploadSpeedOfProgressData(fileHash string, size uint64) *protos.UploadSpeedOfProgress {
	return &protos.UploadSpeedOfProgress{
		FileHash:  fileHash,
		SliceSize: size,
	}
}

func ReqGetCapacityData(reqID string) *protos.ReqGetCapacity {
	return &protos.ReqGetCapacity{
		P2PAddress:    setting.P2PAddress,
		WalletAddress: setting.WalletAddress,
		ReqId:         reqID,
	}
}

func ReqNodeStatusData() *protos.ReqReportNodeStatus {
	// cpu total used percent
	totalPercent, _ := cpu.Percent(3*time.Second, false)
	// num of cpu cores
	coreNum, _ := cpu.Counts(false)
	var cpuPercent float64
	if len(totalPercent) == 0 {
		cpuPercent = 0
	} else {
		cpuPercent = totalPercent[0]
	}
	cpuStat := &protos.CpuStat{NumCores: int64(coreNum), TotalUsedPercent: math.Round(cpuPercent*100) / 100}

	// Memory physical + swap
	virtualMem, _ := mem.VirtualMemory()
	virtualUsedMem := virtualMem.Used
	virtualTotalMem := virtualMem.Total

	swapMemory, _ := mem.SwapMemory()
	swapUsedMem := swapMemory.Used
	swapTotalMem := swapMemory.Total
	memStat := &protos.MemoryStat{
		MemUsed: int64(virtualUsedMem), MemTotal: int64(virtualTotalMem),
		SwapMemUsed: int64(swapUsedMem), SwapMemTotal: int64(swapTotalMem),
	}
	// Disk root path
	info, _ := disk.Usage("/")
	diskUsedRoot := info.Used
	diskTotalRoot := info.Total
	diskStat := &protos.DiskStat{RootUsed: int64(diskUsedRoot), RootTotal: int64(diskTotalRoot)}

	// TODO Bandwidth
	bwStat := &protos.BandwidthStat{}

	req := &protos.ReqReportNodeStatus{
		P2PAddress: setting.P2PAddress,
		Cpu:        cpuStat,
		Memory:     memStat,
		Disk:       diskStat,
		Bandwidth:  bwStat,
	}
	return req
}

// PPMsgHeader
func PPMsgHeader(data []byte, head string) header.MessageHead {
	return header.MakeMessageHeader(1, uint16(setting.Config.Version), uint32(len(data)), head)

}

func UnmarshalData(ctx context.Context, target interface{}) bool {
	msgBuf := core.MessageFromContext(ctx)
	utils.DebugLog("msgBuf len = ", len(msgBuf.MSGData))
	if err := proto.Unmarshal(msgBuf.MSGData, target.(proto.Message)); err != nil {
		utils.ErrorLog("protobuf Unmarshal error,target =", reflect.TypeOf(target))
		return false
	}
	if _, ok := reflect.TypeOf(target).Elem().FieldByName("Data"); !ok {
		utils.DebugLog("target = ", target)
	} else {
		utils.DebugLog("analyse target")
	}
	return true
}

