package doubao

type AsrDoubaoConfig struct {
	Model      string
	Host       string
	ApiKey     string
	AccessKey  string
	ResourceId string
}

func DefaultConfig() *AsrDoubaoConfig {
	return &AsrDoubaoConfig{
		Model:      "bigmodel",
		Host:       "openspeech.bytedance.com",
		ResourceId: DoubaoModelDuration,
		ApiKey:     "",
		AccessKey:  "",
	}
}
