package handlers

import "github.com/bougou/go-ipmi/pkg/types"

const (
	maxSDRReadBytes = 16
)

// RegisterStorageHandlers adds P0 read-only Storage NetFn handlers to r.
func RegisterStorageHandlers(r *Registry) {
	r.RegisterFunc(types.CommandGetFRUInventoryAreaInfo, handleGetFRUInventoryAreaInfo)
	r.RegisterFunc(types.CommandReadFRUData, handleReadFRUData)
	r.RegisterFunc(types.CommandGetSDRRepoInfo, handleGetSDRRepoInfo)
	r.RegisterFunc(types.CommandGetSDRRepoAllocInfo, handleGetSDRRepoAllocInfo)
	r.RegisterFunc(types.CommandReserveSDRRepo, handleReserveSDRRepo)
	r.RegisterFunc(types.CommandGetSDR, handleGetSDR)
}
