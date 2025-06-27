package types

// https://ccnphfhqs21z.feishu.cn/wiki/FjW6wZmisimNBBkov6OcmfvknVd
type Application struct {
	Name        string `json:"name"`         // 应用名称
	Version     string `json:"version"`      // 应用版本
	CompileTime string `json:"compile_time"` // 编译时间
	IdfVersion  string `json:"idf_version"`  // IDF 版本
	ElfSHA256   string `json:"elf_sha256"`   // ELF SHA
}

type Board struct {
	Type    string `json:"type"`    // 板类型
	Name    string `json:"name"`    // 板名称
	Ssid    string `json:"ssid"`    // Wi-Fi SSID
	Rssi    int32  `json:"rssi"`    // Wi-Fi 信号强度
	Channel int32  `json:"channel"` // Wi-Fi 信道
	Ip      string `json:"ip"`      // IP 地址
	Mac     string `json:"mac"`     // MAC 地址

	Revision string `json:"revision"` // 修订版本
	Carrier  string `json:"carrier"`  // 运营商
	Csq      string `json:"csq"`      // 信号质量
	Iccid    string `json:"icc_id"`   // ICCID
}

type PartitionTable struct {
	Label   string `json:"label"`   // 分区标签
	Type    int32  `json:"type"`    // 分区类型
	Subtype int32  `json:"subtype"` // 分区子类型
	Address uint64 `json:"address"` // 分区地址o
	Size    uint64 `json:"size"`    // 分区大小
}

type Ota struct {
	Label string `json:"label"` // OTA 标签
}

type Device struct {
	// from header
	DeviceId       string `json:"device_id"`       // 设备 ID
	ClientId       string `json:"client_id"`       // 客户端 ID
	UserAgent      string `json:"user_agent"`      // 用户代理
	AcceptLanguage string `json:"accept_language"` // 接受语言

	// from body
	Version       int32  `json:"version"`         // 版本号
	Language      string `json:"language"`        // 语言
	FlashSize     uint64 `json:"flash_size"`      // 闪存大小
	MacAddress    string `json:"mac_address"`     // MAC 地址
	ChipModelName string `json:"chip_model_name"` // 芯片型号名称
	UUID          string `json:"uuid"`            // 设备 UUID

	Application Application      `json:"application"` // 应用信息
	Partitions  []PartitionTable `json:"partitions"`  // 分区表
	Ota         Ota              `json:"ota"`         // OTA 信息
	Board       Board            `json:"board"`       // 板信息

	RawBody string `json:"raw_body"` // 原始数据
}
