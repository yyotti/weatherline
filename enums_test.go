package weatherline

import (
	"fmt"
	"testing"
)

func TestLang_String(t *testing.T) {
	tests := []struct {
		l        Lang
		expected string
	}{
		// TEST0 {{{
		{
			l:        LangEn,
			expected: "en (English)",
		},
		// }}}
		// TEST1 {{{
		{
			l:        LangJa,
			expected: "ja (Japanese)",
		},
		// }}}
		// TEST2 {{{
		{
			l:        LangUnknown,
			expected: "?? (Unknown)",
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := tt.l.String()
			if actual != tt.expected {
				t.Errorf("Expected to get [%s], but got [%s]", tt.expected, actual)
			}
		})
	}
}

func TestLang_Value(t *testing.T) {
	tests := []struct {
		l        Lang
		expected string
	}{
		// TEST0 {{{
		{
			l:        LangEn,
			expected: "en",
		},
		// }}}
		// TEST1 {{{
		{
			l:        LangJa,
			expected: "ja",
		},
		// }}}
		// TEST2 {{{
		{
			l:        LangUnknown,
			expected: "",
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := tt.l.Value()
			if actual != tt.expected {
				t.Errorf("Expected to get [%s], but got [%s]", tt.expected, actual)
			}
		})
	}
}

func TestLangValueOf(t *testing.T) {
	tests := []struct {
		s        string
		expected Lang
	}{
		// TEST0 {{{
		{
			s:        "en",
			expected: LangEn,
		},
		// }}}
		// TEST1 {{{
		{
			s:        "ja",
			expected: LangJa,
		},
		// }}}
		// TEST2 {{{
		{
			s:        "",
			expected: LangUnknown,
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := LangValueOf(tt.s)
			if actual != tt.expected {
				t.Errorf("Expected to get [%s], but got [%s]", tt.expected, actual)
			}
		})
	}
}

func TestUnits_String(t *testing.T) {
	tests := []struct {
		u        Units
		expected string
	}{
		// TEST0 {{{
		{
			u:        UnitsUS,
			expected: "us (Imperial units)",
		},
		// }}}
		// TEST1 {{{
		{
			u:        UnitsSI,
			expected: "si (SI units)",
		},
		// }}}
		// TEST2 {{{
		{
			u:        UnitsUnknown,
			expected: "?? (Unknown)",
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := tt.u.String()
			if actual != tt.expected {
				t.Errorf("Expected to get [%s], but got [%s]", tt.expected, actual)
			}
		})
	}
}

func TestUnits_Value(t *testing.T) {
	tests := []struct {
		u        Units
		expected string
	}{
		// TEST0 {{{
		{
			u:        UnitsUS,
			expected: "us",
		},
		// }}}
		// TEST1 {{{
		{
			u:        UnitsSI,
			expected: "si",
		},
		// }}}
		// TEST2 {{{
		{
			u:        UnitsUnknown,
			expected: "",
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := tt.u.Value()
			if actual != tt.expected {
				t.Errorf("Expected to get [%s], but got [%s]", tt.expected, actual)
			}
		})
	}
}

func TestUnitsValueOf(t *testing.T) {
	tests := []struct {
		s        string
		expected Units
	}{
		// TEST0 {{{
		{
			s:        "us",
			expected: UnitsUS,
		},
		// }}}
		// TEST1 {{{
		{
			s:        "si",
			expected: UnitsSI,
		},
		// }}}
		// TEST2 {{{
		{
			s:        "",
			expected: UnitsUnknown,
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := UnitsValueOf(tt.s)
			if actual != tt.expected {
				t.Errorf("Expected to get [%s], but got [%s]", tt.expected, actual)
			}
		})
	}
}

func TestWeather_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json []byte

		expected    Weather
		expectError bool
	}{
		// TEST0 {{{
		{
			json:        []byte(`"clear-day"`),
			expected:    WeatherClearDay,
			expectError: false,
		},
		// }}}
		// TEST1 {{{
		{
			json:        []byte(`"unknown-weather"`),
			expected:    WeatherUnknown,
			expectError: false,
		},
		// }}}
		// TEST2 {{{
		{
			json:        []byte(`1`),
			expectError: true,
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			var w Weather
			err := w.UnmarshalJSON(tt.json)
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

			if w != tt.expected {
				t.Errorf("Expected to get [%+v], but got [%+v]", tt.expected, err)
			}
		})
	}
}
