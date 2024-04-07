package types

type Point struct {
	Coordinates struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coordinates"`
}

type Localization struct {
	Corner struct {
		UpperLeft  Point `json:"upper-left"`
		UpperRight Point `json:"upper-right"`
		LowerLeft  Point `json:"lower-left"`
		LowerRight Point `json:"lower-right"`
	} `json:"corner"`
	CachedObject `json:"-"`
}
