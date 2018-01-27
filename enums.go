package weatherline

import (
	"encoding/json"
)

// Lang : langパラメータ種別
type Lang int

// Languages
const (
	LangUnknown Lang = iota

	LangEn // English (which is the default)
	LangJa // Japanese
)

var langs = map[string]Lang{
	"en": LangEn,
	"ja": LangJa,
}

func (l Lang) String() string {
	switch l {
	case LangEn:
		return "en (English)"
	case LangJa:
		return "ja (Japanese)"
	default:
		return "?? (Unknown)"
	}
}

// Value : 値を返す
func (l Lang) Value() string {
	for k, v := range langs {
		if v == l {
			return k
		}
	}

	return ""
}

// LangValueOf : 文字列をLang型に変換する
func LangValueOf(str string) Lang {
	if lang, ok := langs[str]; ok {
		return lang
	}

	return LangUnknown
}

// Units : unitsパラメータ種別
type Units int

// Unit
const (
	UnitsUnknown Units = iota

	UnitsUS // Imperial units (the default)
	UnitsSI // SI units
)

var unitss = map[string]Units{
	"us": UnitsUS,
	"si": UnitsSI,
}

func (u Units) String() string {
	switch u {
	case UnitsUS:
		return "us (Imperial units)"
	case UnitsSI:
		return "si (SI units)"
	default:
		return "?? (Unknown)"
	}
}

// Value : 値を返す
func (u Units) Value() string {
	for k, v := range unitss {
		if v == u {
			return k
		}
	}

	return ""
}

// UnitsValueOf : 文字列をUnits型に変換する
func UnitsValueOf(str string) Units {
	if units, ok := unitss[str]; ok {
		return units
	}

	return UnitsUnknown
}

// Weather : 天気種別
type Weather int

// Weathers
const (
	WeatherUnknown Weather = iota

	WeatherClearDay
	WeatherClearNight
	WeatherRain
	WeatherSnow
	WeatherSleet
	WeatherWind
	WeatherFog
	WeatherCloudy
	WeatherPartlyCloudyDay
	WeatherPartlyCloudyNight
)

var weathers = map[string]Weather{
	"clear-day":           WeatherClearDay,
	"clear-night":         WeatherClearNight,
	"rain":                WeatherRain,
	"snow":                WeatherSnow,
	"sleet":               WeatherSleet,
	"wind":                WeatherWind,
	"fog":                 WeatherFog,
	"cloudy":              WeatherCloudy,
	"partly-cloudy-day":   WeatherPartlyCloudyDay,
	"partly-cloudy-night": WeatherPartlyCloudyNight,
}

// UnmarshalJSON : json.Unmarshal のための独自実装
func (w *Weather) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	weather, ok := weathers[s]
	if ok {
		*w = weather
	} else {
		*w = WeatherUnknown
	}

	return nil
}
