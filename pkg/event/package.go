package event

type ShortUrlEvent struct {
	ShortUrl string `json:"short_url"`
	LongUrl  string `json:"long_url"`
}

type KafkaConfigs struct {
	BootstrapServers string
	Topic            string
	GroupId          string
	Offset           string
}
