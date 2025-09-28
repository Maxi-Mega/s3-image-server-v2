package types //nolint: revive,nolintlint

type RawFeaturesFile struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
	} `json:"features"`
}

type Features struct {
	CachedObject

	Class   string         `json:"class"`
	Count   int            `json:"featuresCount"`
	Objects map[string]int `json:"objects"`
}
