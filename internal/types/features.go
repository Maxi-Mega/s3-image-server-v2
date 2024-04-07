package types

type RawFeaturesFile struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
	} `json:"features"`
}

type Features struct {
	Class        string         `json:"class"`
	Count        int            `json:"featuresCount"`
	Objects      map[string]int `json:"objects"`
	CachedObject `json:"-"`
}
