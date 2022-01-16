package ipmi

import (
	"fmt"
	"strings"
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
	MaxPrivLevel PrivilegeLevel
}

func (req *GetUserAccessRequest) Command() Command {
	return CommandGetUserAccess
}

func (req *GetUserAccessRequest) Pack() []byte {
	return []byte{req.ChannelNumber, req.UserID}
}

func (res *GetUserAccessResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetUserAccessResponse) Unpack(msg []byte) error {
	if len(msg) < 4 {
		return ErrUnpackedDataTooShort
	}

	res.MaxUsersIDCount, _, _ = unpackUint8(msg, 0)

	b1, _, _ := unpackUint8(msg, 1)
	res.EnableStatus = b1 & 0xc0 >> 6
	res.EnabledUserIDsCount = b1 & 0x3f

	b2, _, _ := unpackUint8(msg, 2)
	res.FixedNameUseIDsCount = b2 & 0x3f

	b3, _, _ := unpackUint8(msg, 3)
	res.CallbackOnly = isBit6Set(b3)
	res.LinkAuthEnabled = isBit5Set(b3)
	res.IPMIMessagingEnabled = isBit4Set(b3)
	res.MaxPrivLevel = PrivilegeLevel(b3 & 0x0f)
	return nil
}

func (res *GetUserAccessResponse) Format() string {
	return ""
}

func (c *Client) GetUserAccess(channelNumber uint8, userID uint8) (response *GetUserAccessResponse, err error) {
	request := &GetUserAccessRequest{
		ChannelNumber: channelNumber,
		UserID:        userID,
	}
	response = &GetUserAccessResponse{}
	err = c.Exchange(request, response)
	return
}

func (c *Client) ListUser(channelNumber uint8) ([]*User, error) {
	var users = make([]*User, 0)

	var userID uint8 = 1
	var username string
	for {
		res, err := c.GetUserAccess(channelNumber, userID)
		if err != nil {
			return nil, fmt.Errorf("get user for userID %d failed, err: %s", userID, err)
		}

		res2, err := c.GetUsername(userID)
		if err != nil {
			respErr, ok := err.(*ResponseError)
			if !ok || uint8(respErr.CompletionCode()) != 0xcc {
				return nil, fmt.Errorf("get user name for userID %d failed, err: %s", userID, err)
			}

			// Completion Code is 0xcc, means this UserID is not set.
			username = ""
		} else {
			username = res2.Username
		}

		user := &User{
			ID:                   userID,
			Name:                 username,
			Callin:               !res.CallbackOnly,
			LinkAuthEnabled:      res.LinkAuthEnabled,
			IPMIMessagingEnabled: res.IPMIMessagingEnabled,
			MaxPrivLevel:         res.MaxPrivLevel,
		}
		users = append(users, user)

		if userID >= res.MaxUsersIDCount {
			break
		}
		userID += 1
	}

	return users, nil
}

type User struct {
	ID                   uint8
	Name                 string
	Callin               bool
	LinkAuthEnabled      bool
	IPMIMessagingEnabled bool
	MaxPrivLevel         PrivilegeLevel
}

func FormatUsers(users []*User) string {
	var lines []string

	headers := []formatValue{
		fv("%-2s", "ID"),
		fv("%-20s", "Name"),
		fv("%-6s", "Callin"),
		fv("%-9s", "Link Auth"),
		fv("%-8s", "IPMI Msg"),
		fv("%-18s", "Channel Priv Limit"),
	}

	lines = append(lines, formatValuesTable(headers))

	for _, user := range users {
		content := []formatValue{
			fv("%-2s", fmt.Sprintf("%d", user.ID)),
			fv("%-20s", user.Name),
			fv("%-6s", fmt.Sprintf("%v", user.Callin)),
			fv("%-9s", fmt.Sprintf("%v", user.LinkAuthEnabled)),
			fv("%-8s", fmt.Sprintf("%v", user.IPMIMessagingEnabled)),
			fv("%-18s", user.MaxPrivLevel),
		}
		lines = append(lines, formatValuesTable(content))
	}

	lines = append(lines, formatValuesTable(headers))

	return strings.Join(lines, "\n")
}
