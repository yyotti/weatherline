package weatherline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	forecastAPIBase = "https://api.darksky.net"
)

type forecastError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

func (e forecastError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

var (
	excludes = []string{
		"currently",
		"minutely",
		"alerts",
		"flags",
	}
)

type timeZone time.Location

// UnmarshalJSON : json.Unmarshal のための独自実装
func (tz *timeZone) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	loc, err := time.LoadLocation(s)
	if err != nil {
		loc = time.UTC
	}

	*tz = timeZone(*loc)

	return nil
}

type apiTime struct {
	time.Time
}

// UnmarshalJSON : json.Unmarshal のための独自実装
func (t *apiTime) UnmarshalJSON(b []byte) error {
	var unixTime int64
	if err := json.Unmarshal(b, &unixTime); err != nil {
		return err
	}

	*t = apiTime{time.Unix(unixTime, 0)}

	return nil
}

// ForecastResponse : 天気情報
type ForecastResponse struct {
	TimeZone timeZone  `json:"timezone"`
	Hourly   dataBlock `json:"hourly"`
	Daily    dataBlock `json:"daily"`
}

type dataBlock struct {
	Data    []dataPoint `json:"data"`
	Icon    Weather     `json:"icon"`
	Summary string      `json:"summary"`
}

type dataPoint struct {
	Weather                     Weather `json:"icon"`
	ApparentTemperature         float64 `json:"apparentTemperature"`         // 体感気温, not on daily
	ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh"`     // 最高体感気温, only on daily
	ApparentTemperatureHighTime apiTime `json:"apparentTemperatureHighTime"` // 最高体感気温時刻, only on daily
	ApparentTemperatureLow      float64 `json:"apparentTemperatureLow"`      // 最低体感気温, only on daily
	ApparentTemperatureLowTime  apiTime `json:"apparentTemperatureLowTime"`  // 最低体感気温時刻, only on daily
	PrecipAccumulation          float64 `json:"precipAccumulation"`          // 積雪量、天気が雪でなければ無視, only on hourly and daily
	PrecipProbability           float64 `json:"precipProbability"`           // 降水確率
	Summary                     string  `json:"summary"`
	Time                        apiTime `json:"time"`
	Temperature                 float64 `json:"temperature"`         // not in minutely
	TemperatureHigh             float64 `json:"temperatureHigh"`     // only on daily
	TemperatureHighTime         apiTime `json:"temperatureHighTime"` // only on daily
	TemperatureLow              float64 `json:"temperatureLow"`      // only on daily
	TemperatureLowTime          apiTime `json:"temperatureLowTime"`  // only on daily
}

// Forecast : forecast API (Dark Sky API) client interface
type Forecast interface {
	Get(Lang, Units) (*ForecastResponse, error)
}

type forecast struct {
	url        *url.URL
	httpClient *http.Client
}

// NewForecast : Create Forecast instance
func NewForecast(token, lat, long string) Forecast {
	u, err := url.Parse(forecastAPIBase)
	if err != nil {
		return nil
	}

	u.Path = path.Join(u.Path, "forecast", token, fmt.Sprintf("%s,%s", lat, long))
	return &forecast{
		url:        u,
		httpClient: &http.Client{},
	}
}

// Get : ForecastClient.Get の実装
func (f *forecast) Get(lang Lang, units Units) (*ForecastResponse, error) {
	values := url.Values{}
	if lang != LangUnknown {
		values.Set("lang", lang.Value())
	}
	if units != UnitsUnknown {
		values.Set("units", units.Value())
	}
	values.Set("exclude", strings.Join(excludes, ","))

	u := *f.url
	u.RawQuery = values.Encode()

	res, err := f.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusOK:
		r := ForecastResponse{}
		err = json.Unmarshal(body, &r)

		return &r, err

	case 400:
		e := forecastError{}
		err = json.Unmarshal(body, &e)
		if err != nil {
			return nil, err
		}

		return nil, e

	default:
		e := forecastError{
			Code:    res.StatusCode,
			Message: string(body),
		}
		return nil, e
	}
}
