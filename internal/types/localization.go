package types //nolint: revive,nolintlint

type Point struct {
	Coordinates struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coordinates"`
}

type LocalizationCorner struct {
	UpperLeft  Point `json:"upper-left"`
	UpperRight Point `json:"upper-right"`
	LowerLeft  Point `json:"lower-left"`
	LowerRight Point `json:"lower-right"`
}

type Localization struct {
	CachedObject

	Corner LocalizationCorner `json:"corner"`
}
