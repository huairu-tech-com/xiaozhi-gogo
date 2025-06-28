package repo

import (
	"github.com/huairu-tech-com/xiaozhi-gogo/pkg/types"
)

// Repository interface defines the methods for device repository operations.
type Respository interface {
	deviceRepo // deviceRepo defines the methods for device operations.
}

type WhereCondition map[string]any

func (wc WhereCondition) MatchDeviceId(deviceId string) bool {
	if wc == nil {
		return true
	}

	if id, ok := wc["device_id"]; ok {
		return id == deviceId
	}

	return false
}

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
