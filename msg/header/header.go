package header

// Author j & cc
import (
	"github.com/stratosnet/sds/utils"
)

// MessageHead
type MessageHead struct {
	Tag     int16
	Len     uint32
	Cmd     []byte //8 byte
	Version uint16
}

// MakeMessageHeader
func MakeMessageHeader(tag int16, varsion uint16, length uint32, cmd string) MessageHead {
	var cmdByte = []byte(cmd)[:8]
	return MessageHead{
		Tag:     tag,
		Len:     length,
		Cmd:     cmdByte,
		Version: varsion,
	}
}

// GetMessageHeader
func GetMessageHeader(tag int16, varsion uint16, length uint32, cmd string, data []byte) {
	var cmdByte = []byte(cmd)[:8]
	copy(data[0:2], utils.Int16ToBytes(tag))
	copy(data[2:6], utils.Uint32ToBytes(length))
	copy(data[6:14], cmdByte)
	copy(data[14:16], utils.Uint16ToBytes(varsion))
}

//cmd, 8 bytes string, exceeded will be truncate
const (
	ReqGetPPList = "ReqGPPL" // request to get pp list
	RspGetPPList = "RspGPPL" // response to get pp list
	ReqGetSPList = "ReqGSPL" // request to get pp list
	RspGetSPList = "RspGSPL" // response to get pp list
	ReqRegister  = "ReqReg"  // request to register
	RspRegister  = "RspReg"  // response to register

	ReqActivatePP  = "ReqActvp" // request to activate a PP node
	RspActivatePP  = "RspActvp" // response to activate a PP node
	ReqActivatedPP = "ReqActdp" // request when a PP node was successfully activated
	ReqActivatedSP = "ReqActds" // request when a SP node was successfully activated
	RspActivatedPP = "RspActdp" // response when a PP node was successfully activated

	ReqDeactivatePP  = "ReqDctvp" // request to deactivate a PP node
	RspDeactivatePP  = "RspDctvp" // response to deactivate a PP node
	ReqDeactivatedPP = "ReqDctdp" // request when a PP node was successfully deactivated
	RspDeactivatedPP = "RspDctdp" // response when a PP node was successfully deactivated

	ReqPrepay  = "ReqPrpay" // request for a PP node sending a prepay transaction
	RspPrepay  = "RspPrpay" // response for a PP node sending a prepay transaction
	ReqPrepaid = "ReqPrpad" // request when a PP node prepay transaction was successful
	RspPrepaid = "RspPrpad" // response when a PP node prepay transaction was successful

	ReqMining = "ReqMin" // request to mining
	RspMining = "RspMin" //  response to mining

	// upload
	ReqUploadFile              = "ReqUpl"
	RspUploadFile              = "RspUpl"
	ReqUploadFileSlice         = "ReqUpLFS"
	RspUploadFileSlice         = "RspUpLFS"
	ReqReportUploadSliceResult = "ReqUFR"
	RspReportUploadSliceResult = "RspUFR"
	UploadSpeedOfProgress      = "USOP"
	Uploaded                   = "Uploaded"

	ReqFindMyFileList = "ReqFFL"
	RspFindMyFileList = "RspFFL"

	ReqFindDirectory = "ReqFDr"
	RspFindDirectory = "RspFDr"

	ReqDeleteFile = "ReqDF"
	RspDeleteFile = "RspDF"

	ReqGetHDInfo = "ReqHDI"
	RspGetHDInfo = "RspHDI"

	ReqDeleteSlice = "ReqDLS"
	RspDeleteSlice = "RspDLS"

	ReqMakeDirectory     = "ReqMDR"
	RspMakeDirectory     = "RspMDR"
	ReqRemoveDirectory   = "ReqRD"
	RspRemoveDirectory   = "RsqRD"
	ReqMoveFileDirectory = "ReqMFD"
	RspMoveFileDirectory = "RspMFD"

	//  download
	ReqFileStorageInfo      = "ReqQDLF"
	RspFileStorageInfo      = "RspQDLF"
	ReqDownloadSlice        = "ReqDLFS"
	RspDownloadSlice        = "RspDLFS"
	ReqReportDownloadResult = "ReqDLRep" // request to download result report
	RspReportDownloadResult = "RspDLRep" // response to download result report
	ReqDownloadTaskInfo     = "ReqDLTI"
	RspDownloadTaskInfo     = "RspDLTI"
	ReqDownloadSliceWrong   = "ReqDSW"
	RspDownloadSliceWrong   = "RspDSW"
	ReqDownloadSlicePause   = "ReqDSP"
	RspDownloadSlicePause   = "RspDSP"
	ReqFindDirectoryTree    = "ReqFDT"
	RspFindDirectoryTree    = "RspFDT"
	ReqClearDownloadTask    = "ReqCDT"

	// register new pp
	ReqRegisterNewPP = "ReqRgNPP"
	RspRegisterNewPP = "RspRgNPP"

	// transfer
	ReqTransferNotice         = "ReqTrNot"
	RspTransferNotice         = "RspTrNot"
	ReqValidateTransferCer    = "ReqVTCer" // request to validate transfer certificate PP->SP
	RspValidateTransferCer    = "RspVTCer" // response to validate transfer certificate SP->PP
	ReqTransferDownload       = "ReqTdl"
	RspTransferDownload       = "RspTdl"
	RspTransferDownloadResult = "RspTdlR"

	ReqReportTransferResult = "ReqTrRep"
	RspReportTransferResult = "RspTrRep"

	//TODO change to report to SP
	ReqReportTaskBP = "ReqRTBP" // report to BP

	ReqShareLink    = "ReqSL"
	RspShareLink    = "RspSL"
	ReqShareFile    = "ReqSF"
	RspShareFile    = "RspSF"
	ReqDeleteShare  = "ReqDSF"
	RspDeleteShare  = "RspDSF"
	ReqGetShareFile = "ReqGSF"
	RspGetShareFile = "RspGSF"

	ReqSaveFile   = "ReqTSF"
	RspSaveFile   = "RspTSF"
	ReqSaveFolder = "ReqTSFD"
	RspSaveFolder = "RspTSFD"

	ReqCreateAlbum       = "ReqCA"
	RspCreateAlbum       = "RspCA"
	ReqEditAlbum         = "ReqEDA"
	RspEditAlbum         = "RspEDA"
	ReqAlbumContent      = "ReqACT"
	RspAlbumContent      = "RspACT"
	ReqSearchAlbum       = "ReqSCA"
	RspSearchAlbum       = "RspSCA"
	ReqFindMyAlbum       = "ReqFMA"
	RspFindMyAlbum       = "RspFMA"
	ReqCollectionAlbum   = "ReqCLA" // request to collect album
	RspCollectionAlbum   = "RspCLA"
	ReqAbstractAlbum     = "ReqAA"
	RspAbstractAlbum     = "RspAA"
	ReqMyCollectionAlbum = "ReqMCA"
	RspMyCollectionAlbum = "RspMCA"
	ReqDeleteAlbum       = "ReqDA"
	RspDeleteAlbum       = "RspDA"
	ReqGetCapacity       = "ReqGC"
	RspGetCapacity       = "RspGC"

	ReqFileSort = "ReqFS"
	RspFileSort = "RspFS"

	ReqConfig = "ReqCon"
	RspConfig = "RspCon"

	ReqInvite    = "ReqInv"
	RspInvite    = "RspInv"
	ReqGetReward = "ReqGR"
	RspGetReward = "RspGR"

	// heartbeat
	ReqHeart = "ReqHeart"
	RspHeart = "RspHeart"

	// report node status
	ReqReportNodeStatus = "ReqRNS"
	RspReportNodeStatus = "RspRNS"

	// customer volume
	ReqCAddVolume = "ReqCAV"
	RspCAddVolume = "RspCAV"

	ReqCUseVolume = "ReqCUV"
	RspCUseVolume = "RspCUV"

	ReqTransferAggregateTraffic = "ReqTrAgT"
	ReqTransferRecordTraffic    = "ReqTrRcT"
)

// DecodeHeader
func DecodeHeader(packet []byte) MessageHead {
	var msgH = MessageHead{
		Tag:     utils.BytesToInt16(packet[:2]),
		Len:     utils.BytesToUInt32(packet[2:6]),
		Cmd:     packet[6:14],
		Version: utils.BytesToUint16(packet[14:]),
	}
	return msgH
}

// NewDecodeHeader
func NewDecodeHeader(packet []byte, msgH *MessageHead) {
	msgH.Tag = utils.BytesToInt16(packet[:2])
	msgH.Len = utils.BytesToUInt32(packet[2:6])
	msgH.Cmd = packet[6:14]
	msgH.Version = utils.BytesToUint16(packet[14:])
}
