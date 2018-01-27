package weatherline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestForecastError_Error(t *testing.T) {
	err := forecastError{
		Code:    100,
		Message: "This is test",
	}

	expected := "100: This is test"

	actual := err.Error()
	if actual != expected {
		t.Errorf("Expected to get [%s], but got [%s]", expected, actual)
	}
}

func loadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return loc
}

func TestTimeZone_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json []byte

		expected    timeZone
		expectError bool
	}{
		// TEST0 {{{
		{
			json:        []byte(`"Asia/Tokyo"`),
			expected:    timeZone(*loadLocation("Asia/Tokyo")),
			expectError: false,
		},
		// }}}
		// TEST1 {{{
		{
			json:        []byte(`"unknown-location"`),
			expected:    timeZone(*time.UTC),
			expectError: false,
		},
		// }}}
		// TEST2 {{{
		{
			json:        []byte(`0`),
			expectError: true,
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			var tz timeZone
			err := tz.UnmarshalJSON(tt.json)
			if err != nil {
				if !tt.expectError {
					t.Errorf("Expected no error occurred, but it occurred (%v)", err)
				}
				return
			}

			if tt.expectError {
				t.Errorf("It was expected that an error occurred, but it did not occur")
				return
			}

			if !reflect.DeepEqual(tz, tt.expected) {
				t.Errorf("Expected to get [%+v], but got [%+v]", tt.expected, err)
			}
		})
	}
}

func TestAPITime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json []byte

		expected    apiTime
		expectError bool
	}{
		// TEST0 {{{
		{
			json:        []byte(`1517138160`),
			expected:    apiTime{time.Unix(1517138160, 0)}, // 2018/01/01
			expectError: false,
		},
		// }}}
		// TEST1 {{{
		{
			json:        []byte(`0`),
			expected:    apiTime{time.Unix(0, 0)}, // 1970/01/01
			expectError: false,
		},
		// }}}
		// TEST2 {{{
		{
			json:        []byte(`""`),
			expectError: true,
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			var tim apiTime
			err := tim.UnmarshalJSON(tt.json)
			if err != nil {
				if !tt.expectError {
					t.Errorf("Expected no error occurred, but it occurred (%v)", err)
				}
				return
			}

			if tt.expectError {
				t.Errorf("It was expected that an error occurred, but it did not occur")
				return
			}

			if !reflect.DeepEqual(tim, tt.expected) {
				t.Errorf("Expected to get [%+v], but got [%+v]", tt.expected, err)
			}
		})
	}
}

func TestNewForecast(t *testing.T) {
	token := "abcde"
	lat := "123.45"
	long := "67.890"

	fore := NewForecast(token, lat, long)

	if fore == nil {
		t.Fatal("function returns nil")
	}

	f, ok := fore.(*forecast)
	if !ok {
		t.Fatal("Expected lineNotify instance, but not.")
	}

	expectedURL := fmt.Sprintf("%s/forecast/%s/%s,%s", forecastAPIBase, token, lat, long)
	if f.url == nil {
		t.Fatal("url is nil")
	} else if f.url.String() != expectedURL {
		t.Fatalf("Expected url is %s, but it's %s.", expectedURL, f.url.String())
	}

	if f.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
}

func matchRegexp(r string, str string) map[string]string {
	reg := regexp.MustCompile(r)
	match := reg.FindStringSubmatch(str)

	m := make(map[string]string)
	if len(match) == 0 {
		return m
	}

	for i, name := range reg.SubexpNames() {
		if i != 0 && name != "" {
			m[name] = match[i]
		}
	}

	return m
}

func readFile(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func writeForecastErrorResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	if _, err := fmt.Fprintf(w, `{"code":%d,"error":"%s"}`, status, message); err != nil {
		panic(err)
	}
}

func forecastFunc(lang Lang, units Units, resStatus int, response string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeForecastErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Unexpected request: method = %s", r.Method))
			return
		}

		url := r.URL

		err := checkQuery(lang, units, url.Query())
		if err != nil {
			writeForecastErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if resStatus == http.StatusOK {
			w.WriteHeader(resStatus)
			w.Write([]byte(response))
			return
		}

		writeForecastErrorResponse(w, resStatus, response)
	}
}

func checkQuery(lang Lang, units Units, query neturl.Values) error {
	val := query.Get("lang")
	if val != lang.Value() {
		return fmt.Errorf("Unexpected request: `lang` in query = %s", val)
	}

	if val := query.Get("units"); val != units.Value() {
		return fmt.Errorf("Unexpected request: `units` in query = %s", val)
	}

	if val := query.Get("exclude"); val != strings.Join(excludes, ",") {
		return fmt.Errorf("Unexpected request: `exclude` in query = %s", val)
	}

	return nil
}

func unmarshal(str string) *ForecastResponse {
	r := &ForecastResponse{}
	err := json.Unmarshal([]byte(str), r)
	if err != nil {
		panic(err)
	}
	return r
}

func TestForecast_Send(t *testing.T) {
	tests := []struct {
		lang  Lang
		units Units

		resStatus  int
		resMessage string

		expectedResponse *ForecastResponse
		expectedError    error
	}{
		// TEST0 {{{
		{
			lang:  LangJa,
			units: UnitsSI,

			resStatus:  http.StatusOK,
			resMessage: readFile("testdata/forecast/get00.json"),

			expectedResponse: unmarshal(readFile("testdata/forecast/get00.json")),
			expectedError:    nil,
		},
		// }}}
		// TEST1 {{{
		{
			lang:  LangEn,
			units: UnitsUS,

			resStatus:  http.StatusOK,
			resMessage: readFile("testdata/forecast/get01.json"),

			expectedResponse: unmarshal(readFile("testdata/forecast/get01.json")),
			expectedError:    nil,
		},
		// }}}
		// TEST2 {{{
		{
			lang:  LangJa,
			units: UnitsSI,

			resStatus:  400,
			resMessage: "This error is expected",

			expectedError: forecastError{
				Code:    400,
				Message: "This error is expected",
			},
		},
		// }}}
		// TEST3 {{{
		{
			resStatus:  http.StatusInternalServerError,
			resMessage: "This error is expected 2",

			expectedError: forecastError{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf(`{"code":%d,"error":"%s"}`, http.StatusInternalServerError, "This error is expected 2"),
			},
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			server := httptest.NewTLSServer(http.HandlerFunc(forecastFunc(tt.lang, tt.units, tt.resStatus, tt.resMessage)))
			defer server.Close()

			var err error
			f := &forecast{}
			f.url, err = neturl.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			f.httpClient = server.Client()

			res, err := f.Get(tt.lang, tt.units)
			if err != nil {
				if tt.expectedError == nil {
					t.Errorf("Expected no error occurred, but it occurred (%v)", err)
				} else if err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected to get [%v], but got [%v]", tt.expectedError, err)
				}
			} else {
				if tt.expectedError != nil {
					t.Errorf("It was expected that an error occurred, but it did not occur")
				} else if !reflect.DeepEqual(res, tt.expectedResponse) {
					t.Errorf("Expected to get [%+v], but got [%+v]", tt.expectedResponse, res)
				}
			}
		})
	}
}
