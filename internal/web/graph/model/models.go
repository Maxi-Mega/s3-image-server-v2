package model

type FileSelector struct {
	Regex string `json:"regex"`
	Kind  string `json:"kind"`
	Link  bool   `json:"link"`
}

type DynamicData struct {
	FileSelectors map[string]FileSelector `json:"fileSelectors"`
	Expressions   map[string]string       `json:"expressions"`
}
