package repo

import (
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"
)

type WhereCondition map[string]interface{}

func (wc WhereCondition) MatchDevice(d *types.Device) bool {
	if wc == nil {
		return true
	}

	if deviceId, ok := wc["device_id"]; ok {
		return deviceId == d.DeviceId
	}

	if clientId, ok := wc["client_id"]; ok {
		return clientId == d.ClientId
	}

	return false
}

type Respository interface {
	deviceRepo
}
