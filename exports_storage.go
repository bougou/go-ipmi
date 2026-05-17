package ipmi

import storage "github.com/bougou/go-ipmi/pkg/cmd/storage"

type (
	AddSELEntryRequest              = storage.AddSELEntryRequest
	AddSELEntryResponse             = storage.AddSELEntryResponse
	ClearSELRequest                 = storage.ClearSELRequest
	ClearSELResponse                = storage.ClearSELResponse
	DeleteSELEntryRequest           = storage.DeleteSELEntryRequest
	DeleteSELEntryResponse          = storage.DeleteSELEntryResponse
	GetDeviceSDRInfoRequest         = storage.GetDeviceSDRInfoRequest
	GetDeviceSDRInfoResponse        = storage.GetDeviceSDRInfoResponse
	GetDeviceSDRRequest             = storage.GetDeviceSDRRequest
	GetDeviceSDRResponse            = storage.GetDeviceSDRResponse
	GetFRUInventoryAreaInfoRequest  = storage.GetFRUInventoryAreaInfoRequest
	GetFRUInventoryAreaInfoResponse = storage.GetFRUInventoryAreaInfoResponse
	GetSDRRepoAllocInfoRequest      = storage.GetSDRRepoAllocInfoRequest
	GetSDRRepoAllocInfoResponse     = storage.GetSDRRepoAllocInfoResponse
	GetSDRRepoInfoRequest           = storage.GetSDRRepoInfoRequest
	GetSDRRepoInfoResponse          = storage.GetSDRRepoInfoResponse
	GetSDRRequest                   = storage.GetSDRRequest
	GetSDRResponse                  = storage.GetSDRResponse
	GetSELAllocInfoRequest          = storage.GetSELAllocInfoRequest
	GetSELAllocInfoResponse         = storage.GetSELAllocInfoResponse
	GetSELEntryRequest              = storage.GetSELEntryRequest
	GetSELEntryResponse             = storage.GetSELEntryResponse
	GetSELInfoRequest               = storage.GetSELInfoRequest
	GetSELInfoResponse              = storage.GetSELInfoResponse
	GetSELTimeRequest               = storage.GetSELTimeRequest
	GetSELTimeResponse              = storage.GetSELTimeResponse
	GetSELTimeUTCOffsetRequest      = storage.GetSELTimeUTCOffsetRequest
	GetSELTimeUTCOffsetResponse     = storage.GetSELTimeUTCOffsetResponse
	ReadFRUDataRequest              = storage.ReadFRUDataRequest
	ReadFRUDataResponse             = storage.ReadFRUDataResponse
	ReserveDeviceSDRRepoRequest     = storage.ReserveDeviceSDRRepoRequest
	ReserveDeviceSDRRepoResponse    = storage.ReserveDeviceSDRRepoResponse
	ReserveSDRRepoRequest           = storage.ReserveSDRRepoRequest
	ReserveSDRRepoResponse          = storage.ReserveSDRRepoResponse
	ReserveSELRequest               = storage.ReserveSELRequest
	ReserveSELResponse              = storage.ReserveSELResponse
	SDROperationSupport             = storage.SDROperationSupport
	SELOperationSupport             = storage.SELOperationSupport
	SetSELTimeRequest               = storage.SetSELTimeRequest
	SetSELTimeResponse              = storage.SetSELTimeResponse
	SetSELTimeUTCOffsetRequest      = storage.SetSELTimeUTCOffsetRequest
	SetSELTimeUTCOffsetResponse     = storage.SetSELTimeUTCOffsetResponse
	WriteFRUDataRequest             = storage.WriteFRUDataRequest
	WriteFRUDataResponse            = storage.WriteFRUDataResponse
)

var (
	ReadFRUDataLength2Big = storage.ReadFRUDataLength2Big
)
