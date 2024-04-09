package types

import (
	"sort"
	"strings"
)

type GeonamesObject struct {
	Name   string `json:"name"`
	States []struct {
		Name     string `json:"name"`
		Counties []struct {
			Name   string `json:"name"`
			Cities []struct {
				Name string `json:"name"`
			} `json:"cities"`
			Villages []struct {
				Name string `json:"name"`
			} `json:"villages"`
		} `json:"counties"`
	} `json:"states"`
}

type Geonames struct {
	Objects []GeonamesObject `json:"objects"`
	CachedObject
}

func (geonames *Geonames) Sort() {
	sort.Slice(geonames.Objects, func(i, j int) bool {
		return strings.Compare(strings.ToLower(geonames.Objects[i].Name), strings.ToLower(geonames.Objects[j].Name)) < 0
	})

	for o := range geonames.Objects {
		obj := geonames.Objects[o]
		sort.Slice(obj.States, func(i, j int) bool {
			return strings.Compare(strings.ToLower(obj.States[i].Name), strings.ToLower(obj.States[j].Name)) < 0
		})

		for s := range obj.States {
			state := obj.States[s]
			sort.Slice(state.Counties, func(i, j int) bool {
				return strings.Compare(strings.ToLower(state.Counties[i].Name), strings.ToLower(state.Counties[j].Name)) < 0
			})
		}
	}
}

func (geonames *Geonames) GetTopLevel() string {
	if len(geonames.Objects) > 0 { //nolint:nestif
		name := geonames.Objects[0].Name

		states := geonames.Objects[0].States
		if len(states) > 0 {
			name += " / " + states[0].Name

			counties := states[0].Counties
			if len(counties) > 0 {
				name += " / " + counties[0].Name

				cities := counties[0].Cities
				if len(cities) > 0 {
					name += " / " + cities[0].Name
				}
			}
		}

		return name
	}

	return "no geoname found"
}
