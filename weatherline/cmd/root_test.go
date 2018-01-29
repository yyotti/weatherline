package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestRequiredFlagsNotSetError_Error(t *testing.T) {
	tests := []struct {
		err      requiredFlagsNotSetError
		expected string
	}{
		// TEST0 {{{
		{
			err:      requiredFlagsNotSetError{},
			expected: "required flag(s) [] not set",
		},
		// }}}
		// TEST1 {{{
		{
			err:      requiredFlagsNotSetError([]string{}),
			expected: "required flag(s) [] not set",
		},
		// }}}
		// TEST2 {{{
		{
			err:      requiredFlagsNotSetError([]string{"aaa"}),
			expected: `required flag(s) ["aaa"] not set`,
		},
		// }}}
		// TEST3 {{{
		{
			err:      requiredFlagsNotSetError([]string{"aaa", "bbb", "ccc"}),
			expected: `required flag(s) ["aaa","bbb","ccc"] not set`,
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := tt.err.Error()
			if actual != tt.expected {
				t.Errorf("Expected to get %s, but got %s", tt.expected, actual)
			}
		})
	}
}

func TestCheckArgs(t *testing.T) {
	tests := []struct {
		args     []string
		expected error
	}{
		// TEST0 {{{
		{
			args:     []string{},
			expected: nil,
		},
		// }}}
		// TEST1 {{{
		{
			args:     []string{"20180101"},
			expected: nil,
		},
		// }}}
		// TEST2 {{{
		{
			args:     []string{"20180101", "20180102"},
			expected: errTooManyArguments,
		},
		// }}}
		// TEST3 {{{
		{
			args:     []string{"2018/01/01"},
			expected: fmt.Errorf(`parsing time "2018/01/01": month out of range`),
		},
		// }}}
	}

	for i, tt := range tests {
		tt := tt // capture
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			t.Parallel()

			actual := checkArgs(tt.args)
			if actual != nil {
				if tt.expected == nil {
					t.Errorf("Expected no error occurred, but it occurred (%v)", actual)
				} else if actual != tt.expected && actual.Error() != tt.expected.Error() {
					t.Errorf("Expected to get [%#v], but got [%#v]", tt.expected, actual)
				}
			} else {
				if tt.expected != nil {
					t.Errorf("Expected that an error occurred, but it did not occur")
				}
			}
		})
	}
}

func TestCheckConfig(t *testing.T) {
	tests := []struct {
		flags    map[string]string
		expected error
	}{
		// TEST0 {{{
		{
			flags: map[string]string{},
			expected: requiredFlagsNotSetError([]string{
				"line-token",
				"forecast-token",
				"latitude",
				"longitude",
			}),
		},
		// }}}
		// TEST1 {{{
		{
			flags: map[string]string{
				"line-token":     "",
				"forecast-token": "XXXXX",
				"latitude":       "123.45",
				"longitude":      "",
			},
			expected: requiredFlagsNotSetError([]string{
				"line-token",
				"longitude",
			}),
		},
		// }}}
		// TEST2 {{{
		{
			flags: map[string]string{
				"line-token":     "YYYYY",
				"forecast-token": "XXXXX",
				"latitude":       "123.45",
				"longitude":      "67.890",
			},
			expected: nil,
		},
		// }}}
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			viper.Reset()

			for k, v := range tt.flags {
				viper.Set(k, v)
			}

			actual := checkConfig()
			if actual != nil {
				if tt.expected == nil {
					t.Errorf("Expected no error occurred, but it occurred (%v)", actual)
				} else if actual.Error() != tt.expected.Error() {
					t.Errorf("Expected to get [%#v], but got [%#v]", tt.expected, actual)
				}
			} else {
				if tt.expected != nil {
					t.Errorf("Expected that an error occurred, but it did not occur")
				}
			}
		})
	}
}

var errExpected = fmt.Errorf("This is expected error")

func createDummyConfig(toml string) string {
	tempFile, err := ioutil.TempFile(os.TempDir(), "wl")
	if err != nil {
		panic(err)
	}
	defer tempFile.Close()

	if _, err = io.WriteString(tempFile, toml); err != nil {
		panic(err)
	}

	return tempFile.Name()
}

func TestPreRun(t *testing.T) {
	tests := []struct {
		argsFunc   func([]string) error
		configFunc func() error
		configTOML string

		expected error
	}{
		// TEST0 {{{
		{
			expected: nil,
		},
		// }}}
		// TEST1 {{{
		{
			configTOML: "[dummy\nline-token='XXXXX'", // Invalid format
			expected:   fmt.Errorf("While parsing config: (1, 2): unexpected token unclosed table key, was expecting a table key"),
		},
		// }}}
		// TEST2 {{{
		{
			argsFunc: func([]string) error {
				return errExpected
			},
			expected: errExpected,
		},
		// }}}
		// TEST3 {{{
		{
			configFunc: func() error {
				return errExpected
			},
			expected: errExpected,
		},
		// }}}
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			viper.Reset()

			argsFunc := tt.argsFunc
			if argsFunc == nil {
				argsFunc = func([]string) error {
					return nil
				}
			}

			configFunc := tt.configFunc
			if configFunc == nil {
				configFunc = func() error {
					return nil
				}
			}

			if tt.configTOML != "" {
				file := createDummyConfig(tt.configTOML)
				defer os.Remove(file)

				viper.SetConfigType("toml")
				viper.SetConfigFile(file)
			}

			checkArgs = argsFunc
			checkConfig = configFunc

			actual := preRun(nil, []string{"20180101"})
			if actual != nil {
				if tt.expected == nil {
					t.Errorf("Expected no error occurred, but it occurred (%v)", actual)
				} else if actual.Error() != tt.expected.Error() {
					t.Errorf("Expected to get [%#v], but got [%#v]", tt.expected, actual)
				}
			} else {
				if tt.expected != nil {
					t.Errorf("Expected that an error occurred, but it did not occur")
				}
			}
		})
	}
}
