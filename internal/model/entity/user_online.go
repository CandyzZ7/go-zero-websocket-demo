package entity

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
