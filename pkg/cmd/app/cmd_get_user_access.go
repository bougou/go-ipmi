package app

import (
	"fmt"

	"github.com/bougou/go-ipmi/pkg/types"
)

// 22.27 Get User Access Command
type GetUserAccessRequest struct {
	ChannelNumber uint8
	UserID        uint8
}

type GetUserAccessResponse struct {
	// Maximum number of User IDs. 1-based. Count includes User 1. A value of 1
	// indicates only User 1 is supported.
	MaxUsersIDCount uint8

	// [7:6] - User ID Enable status (for IPMI v2.0 errata 3 and later implementations).
	// 00b = User ID enable status unspecified. (For backward compatibility
	// with pre-errata 3 implementations. IPMI errata 3 and later
	// implementations should return the 01b and 10b responses.)
	// 01b = User ID enabled via Set User Password command.
	// 10b = User ID disabled via Set User Password command.
	// 11b = reserved
	EnableStatus uint8

	// [5:0] - count of currently enabled user IDs on this channel (Indicates how
	// many User ID slots are presently in use.)
	EnabledUserIDsCount uint8

	// Count of User IDs with fixed names, including User 1 (1-based). Fixed names
	// in addition to User 1 are required to be associated with sequential user IDs
	// starting from User ID 2.
	FixedNameUseIDsCount uint8

	// [6] - 0b = user access available during call-in or callback direct connection
	//       1b = user access available only during callback connection
	CallbackOnly bool

	// [5] - 0b = user disabled for link authentication
	//       1b = user enabled for link authentication
	LinkAuthEnabled bool

	// [4] - 0b = user disabled for IPMI Messaging
	//       1b = user enabled for IPMI Messaging
	IPMIMessagingEnabled bool

	// [3:0] - User Privilege Limit for given Channel
	MaxPrivLevel types.PrivilegeLevel
}

func (req *GetUserAccessRequest) Command() types.Command {
	return types.CommandGetUserAccess
}

func (req *GetUserAccessRequest) Pack() []byte {
	return []byte{req.ChannelNumber, req.UserID}
}

func (res *GetUserAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetUserAccessResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return types.ErrUnpackedDataTooShortWith(len(msg), 4)
	}

	res.MaxUsersIDCount, _, _ = types.UnpackUint8(msg, 0)

	b1, _, _ := types.UnpackUint8(msg, 1)
	res.EnableStatus = b1 & 0xc0 >> 6
	res.EnabledUserIDsCount = b1 & 0x3f

	b2, _, _ := types.UnpackUint8(msg, 2)
	res.FixedNameUseIDsCount = b2 & 0x3f

	b3, _, _ := types.UnpackUint8(msg, 3)
	res.CallbackOnly = types.IsBit6Set(b3)
	res.LinkAuthEnabled = types.IsBit5Set(b3)
	res.IPMIMessagingEnabled = types.IsBit4Set(b3)
	res.MaxPrivLevel = types.PrivilegeLevel(b3 & 0x0f)
	return nil
}

func (res *GetUserAccessResponse) Format() string {
	return "" +
		fmt.Sprintf("Maximum IDs        : %d\n", res.MaxUsersIDCount) +
		fmt.Sprintf("Enabled User Count : %d\n", res.EnabledUserIDsCount) +
		fmt.Sprintf("Fixed Name Count   : %d\n", res.FixedNameUseIDsCount)
}

// Completion Code is 0xcc, means this UserID is not set.

type User struct {
	ID                   uint8
	Name                 string
	Callin               bool
	LinkAuthEnabled      bool
	IPMIMessagingEnabled bool
	MaxPrivLevel         types.PrivilegeLevel
}

func FormatUsers(users []*User) string {
	rows := make([]map[string]string, len(users))

	for i, user := range users {
		rows[i] = map[string]string{
			"ID":                 fmt.Sprintf("%d", user.ID),
			"Name":               user.Name,
			"Callin":             fmt.Sprintf("%v", user.Callin),
			"Link Auth":          fmt.Sprintf("%v", user.LinkAuthEnabled),
			"IPMI Msg":           fmt.Sprintf("%v", user.IPMIMessagingEnabled),
			"Channel Priv Limit": user.MaxPrivLevel.String(),
		}
	}

	headers := []string{
		"ID",
		"Name",
		"Callin",
		"Link Auth",
		"IPMI Msg",
		"Channel Priv Limit",
	}

	return types.RenderTable(headers, rows)
}
