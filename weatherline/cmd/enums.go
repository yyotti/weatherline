package cmd

import (
	"github.com/yyotti/weatherline"
)

type icon rune

// icons
const (
	iconUnknown icon = 0x2753

	iconClearDay          icon = 0x2600
	iconClearNight        icon = 0x2600
	iconRain              icon = 0x2603
	iconSnow              icon = 0x2744
	iconSleet             icon = 0x2603
	iconWind              icon = 0x1f343
	iconFog               icon = 0x1f32b
	iconCloudy            icon = 0x2601
	iconPartlyCloudyDay   icon = 0x26c5
	iconPartlyCloudyNight icon = 0x26c5
)

var icons = map[weatherline.Weather]icon{
	weatherline.WeatherClearDay:          iconClearDay,
	weatherline.WeatherClearNight:        iconClearNight,
	weatherline.WeatherRain:              iconRain,
	weatherline.WeatherSnow:              iconSnow,
	weatherline.WeatherSleet:             iconSleet,
	weatherline.WeatherWind:              iconWind,
	weatherline.WeatherFog:               iconFog,
	weatherline.WeatherCloudy:            iconCloudy,
	weatherline.WeatherPartlyCloudyDay:   iconPartlyCloudyDay,
	weatherline.WeatherPartlyCloudyNight: iconPartlyCloudyNight,
}
