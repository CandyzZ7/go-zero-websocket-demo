package entity

import (
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	heartbeatTimeout = 3 * 60 // 用户心跳超时时间
)

type UserOnline struct {
	AccIp         string `json:"accIp"`         // acc Ip
	AccPort       string `json:"accPort"`       // acc 端口
	AppID         uint32 `json:"appID"`         // appID
	UserID        string `json:"userID"`        // 用户ID
	ClientIp      string `json:"clientIp"`      // 客户端Ip
	ClientPort    string `json:"clientPort"`    // 客户端端口
	LoginTime     uint64 `json:"loginTime"`     // 用户上次登录时间
	HeartbeatTime uint64 `json:"heartbeatTime"` // 用户上次心跳时间
	LogOutTime    uint64 `json:"logOutTime"`    // 用户退出登录的时间
	Qua           string `json:"qua"`           // qua
	DeviceInfo    string `json:"deviceInfo"`    // 设备信息
	IsLogoff      bool   `json:"isLogoff"`      // 是否下线
}

// Heartbeat 用户心跳
func (m *UserOnline) Heartbeat(currentTime uint64) {
	m.HeartbeatTime = currentTime
	m.IsLogoff = false
}

func (m *UserOnline) Login(accIp, accPort string, appID uint32, userID string, addr string, loginTime uint64) {
	m.AccIp = accIp
	m.AccPort = accPort
	m.AppID = appID
	m.UserID = userID
	m.ClientIp = addr
	m.ClientPort = addr
	m.LoginTime = loginTime
	m.HeartbeatTime = loginTime
	m.IsLogoff = false
}

func (m *UserOnline) Logout() {
	m.LogOutTime = uint64(time.Now().Unix())
	m.IsLogoff = true
}

// IsOnline 判断用户是否在线
func (m *UserOnline) IsOnline() bool {
	if m.IsLogoff {
		return false
	}
	currentTime := uint64(time.Now().Unix())
	if m.HeartbeatTime < (currentTime - heartbeatTimeout) {
		logx.Infof("user heartbeat timeout, appID: %d, userID: %s, heartbeatTime: %d", m.AppID, m.UserID, m.HeartbeatTime)
		return false
	}
	if m.IsLogoff {
		logx.Infof("user is logoff, appID: %d, userID: %s", m.AppID, m.UserID)
		return false
	}
	return true
}

func (m *UserOnline) UserIsLocal(localIp, localPort string) (result bool) {
	if m.AccIp == localIp && m.AccPort == localPort {
		return true
	}
	return false
}
