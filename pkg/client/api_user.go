package client

import (
	"context"
	"fmt"

	"github.com/bougou/go-ipmi/pkg/command/app"
	"github.com/bougou/go-ipmi/pkg/types"
)

func (c *Client) GetUserAccess(ctx context.Context, channelNumber uint8, userID uint8) (response *app.GetUserAccessResponse, err error) {
	request := &app.GetUserAccessRequest{
		ChannelNumber: channelNumber,
		UserID:        userID,
	}
	response = &app.GetUserAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetUsers(ctx context.Context, channelNumber uint8) ([]*app.User, error) {
	var users = make([]*app.User, 0)

	var userID uint8 = 1
	var username string
	for {
		res, err := c.GetUserAccess(ctx, channelNumber, userID)
		if err != nil {
			return nil, fmt.Errorf("GetUserAccess for userID %d failed, err: %w", userID, err)
		}

		res2, err := c.GetUsername(ctx, userID)
		if err != nil {
			if respErr, ok := types.IsResponseError(err); ok {
				if respErr.CompletionCode() == types.CodeRequestDataFieldInvalid {

					username = ""
				}
			} else {
				return nil, fmt.Errorf("GetUsername for userID %d failed, err: %w", userID, err)
			}
		} else {
			username = res2.Username
		}

		user := &app.User{
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

func (c *Client) SetUserAccess(ctx context.Context, request *app.SetUserAccessRequest) (response *app.SetUserAccessResponse, err error) {
	response = &app.SetUserAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetUserPayloadAccess(ctx context.Context, channelNumber uint8, userID uint8) (response *app.GetUserPayloadAccessResponse, err error) {
	request := &app.GetUserPayloadAccessRequest{
		ChannelNumber: channelNumber,
		UserID:        userID,
	}
	response = &app.GetUserPayloadAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetUsername(ctx context.Context, userID uint8) (response *app.GetUsernameResponse, err error) {
	request := &app.GetUsernameRequest{
		UserID: userID,
	}
	response = &app.GetUsernameResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetUserPassword(ctx context.Context, userID uint8, password string, stored20 bool) (response *app.SetUserPasswordResponse, err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Stored20:  stored20,
		Operation: app.PasswordOperationSetPassword,
		Password:  password,
	}
	response = &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) TestUserPassword(ctx context.Context, userID uint8, password string, stored20 bool) (response *app.SetUserPasswordResponse, err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Stored20:  stored20,
		Operation: app.PasswordOperationTestPassword,
		Password:  password,
	}
	response = &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) DisableUser(ctx context.Context, userID uint8) (err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Operation: app.PasswordOperationDisableUser,
	}
	response := &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return err
}

func (c *Client) EnableUser(ctx context.Context, userID uint8) (err error) {
	request := &app.SetUserPasswordRequest{
		UserID:    userID,
		Operation: app.PasswordOperationEnableUser,
	}
	response := &app.SetUserPasswordResponse{}
	err = c.Exchange(ctx, request, response)
	return err
}

func (c *Client) SetUserPayloadAccess(ctx context.Context, payloadType types.PayloadType, payloadInstance uint8) (response *app.SetUserPayloadAccessResponse, err error) {
	request := &app.SetUserPayloadAccessRequest{}
	response = &app.SetUserPayloadAccessResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) SetUsername(ctx context.Context, userID uint8, username string) (response *app.SetUsernameResponse, err error) {
	request := &app.SetUsernameRequest{
		UserID:   userID,
		Username: username,
	}
	response = &app.SetUsernameResponse{}
	err = c.Exchange(ctx, request, response)
	return
}
