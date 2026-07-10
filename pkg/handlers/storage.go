package handlers

// Storage command IDs (spec §33–34).
const (
	CmdGetFRUInventoryAreaInfo uint8 = 0x10
	CmdReadFRUData             uint8 = 0x11
	CmdGetSDRRepoInfo          uint8 = 0x20
	CmdGetSDRRepoAllocInfo     uint8 = 0x21
	CmdReserveSDRRepo          uint8 = 0x22
	CmdGetSDR                  uint8 = 0x23
)

const (
	maxSDRReadBytes = 16
)

// RegisterStorageHandlers adds P0 read-only Storage NetFn handlers to r.
func RegisterStorageHandlers(r *Registry) {
	r.Register(NetFnStorageRequest, CmdGetFRUInventoryAreaInfo, HandlerFunc(handleGetFRUInventoryAreaInfo))
	r.Register(NetFnStorageRequest, CmdReadFRUData, HandlerFunc(handleReadFRUData))
	r.Register(NetFnStorageRequest, CmdGetSDRRepoInfo, HandlerFunc(handleGetSDRRepoInfo))
	r.Register(NetFnStorageRequest, CmdGetSDRRepoAllocInfo, HandlerFunc(handleGetSDRRepoAllocInfo))
	r.Register(NetFnStorageRequest, CmdReserveSDRRepo, HandlerFunc(handleReserveSDRRepo))
	r.Register(NetFnStorageRequest, CmdGetSDR, HandlerFunc(handleGetSDR))
}
