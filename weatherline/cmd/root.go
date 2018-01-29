package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yyotti/weatherline"
)

const (
	vendor        = "yyotti"
	appName       = "weatherline"
	dateArgFormat = "20060102"

	dateRange = 3
)

const (
	configLineToken     = "line-token"
	configForecastToken = "forecast-token"
	configLongitude     = "longitude"
	configLatitude      = "latitude"
	configLang          = "lang"
	configUnits         = "units"
)

var (
	cfgFile string

	version string

	xdgDirs = xdg.New(vendor, appName)

	errTooManyArguments = fmt.Errorf("Too many arguments")

	examples = []string{
		fmt.Sprintf("  %s --line-token=XXXXX --forecast-token=YYYYY 20180101   # Send forecast on 2018/01/01", appName),
		fmt.Sprintf("  %s --line-token=XXXXX --forecast-token=YYYYY --lang=ja  # Send forecast on today by Japanese", appName),
	}

	today = truncHour(time.Now())

	lineNotify weatherline.LineNotify
	forecast   weatherline.Forecast
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "weatherline",
	Short:   "Send weather forecast to LINE",
	Long:    `Get weather forecast from Forecast (Dark Sky) API and send it by LINE Notify API`,
	Example: strings.Join(examples, "\n"),
	Version: version,
	PreRunE: preRun,
	RunE:    run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	if rootCmd.Version == "" {
		rootCmd.Version = "develop"
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	rootCmd.PersistentFlags().StringP(configLineToken, "L", "", "API token for LINE Notify API")
	rootCmd.PersistentFlags().StringP(configForecastToken, "F", "", "API token for Forecast (Dark Sky) API")
	rootCmd.PersistentFlags().StringP(configLongitude, "x", "", "longitude")
	rootCmd.PersistentFlags().StringP(configLatitude, "y", "", "latitude")
	rootCmd.PersistentFlags().StringP(configLang, "l", weatherline.LangEn.Value(),
		fmt.Sprintf("language [%s|%s]", weatherline.LangEn.Value(), weatherline.LangJa.Value()))
	rootCmd.PersistentFlags().StringP(configUnits, "u", weatherline.UnitsUS.Value(),
		fmt.Sprintf("language [%s|%s]", weatherline.UnitsUS.Value(), weatherline.UnitsSI.Value()))

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// config file name
		viper.SetConfigName(appName)

		// search paths
		// priorities are as follows:

		//   1. Current directory
		viper.AddConfigPath(".")

		//   2. Exe directory
		exe, err := os.Executable()
		if err == nil {
			viper.AddConfigPath(filepath.Dir(exe))
		}

		//   3. Paths according to XDG Base Directory Specification
		//      3-1. $XDG_CONFIG_HOME/[vendor]/[appName]
		//      3-2. $XDG_CONFIG_DIRS/[vendor]/[appName]
		viper.AddConfigPath(xdgDirs.ConfigHome())
		for _, d := range xdgDirs.ConfigDirs() {
			viper.AddConfigPath(d)
		}
	}
}

type requiredFlagsNotSetError []string

func (e requiredFlagsNotSetError) Error() string {
	names := []string{}
	for _, n := range e {
		names = append(names, fmt.Sprintf(`"%s"`, n))
	}
	return fmt.Sprintf("required flag(s) [%s] not set", strings.Join(names, ","))
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			// OK
		default:
			return err
		}
	}

	if err := checkConfig(); err != nil {
		return err
	}

	if err := checkArgs(args); err != nil {
		return err
	}

	lineNotify = weatherline.NewLineNotify(viper.GetString(configLineToken))
	forecast = weatherline.NewForecast(viper.GetString(configForecastToken), viper.GetString(configLatitude), viper.GetString(configLongitude))

	return nil
}

var checkConfig = func() error {
	err := requiredFlagsNotSetError{}
	for _, f := range []string{"line-token", "forecast-token", "latitude", "longitude"} {
		switch f {
		default:
			if viper.GetString(f) == "" {
				err = append(err, f)
			}
		}
	}

	if len(err) > 0 {
		return err
	}

	// TODO Check lang/units

	return nil
}

var checkArgs = func(args []string) error {
	if len(args) == 0 {
		return nil
	}
	if len(args) > 1 {
		return errTooManyArguments
	}

	_, err := time.Parse(dateArgFormat, args[0])
	return err
}

func run(cmd *cobra.Command, args []string) error {
	date := today
	if len(args) > 0 {
		var err error
		date, err = time.Parse(dateArgFormat, args[0])
		if err != nil {
			return err
		}
	}

	lang := weatherline.LangValueOf(viper.GetString("lang"))
	units := weatherline.UnitsValueOf(viper.GetString("units"))

	f, err := forecast.Get(lang, units)
	if err != nil {
		return err
	}

	tz := time.Location(f.TimeZone)
	date = truncHour(date.In(&tz))

	var buf bytes.Buffer

	buf.WriteString("\n")
	buf.WriteString(date.Format("01/02"))
	buf.WriteString("\n")

	hourly := createHourly(date, f)
	if hourly != "" {
		buf.WriteString(hourly)
		buf.WriteString("\n")
	}

	daily := createDaily(date, f)
	if daily != "" {
		buf.WriteString(daily)
	}

	return lineNotify.Send(buf.String())
}

func createHourly(date time.Time, f *weatherline.ForecastResponse) string {
	if len(f.Hourly.Data) == 0 {
		return ""
	}

	var buf bytes.Buffer
	for _, point := range f.Hourly.Data {
		d := truncHour(point.Time.Time)
		if !d.Equal(date) {
			continue
		}

		buf.WriteString("  ")
		buf.WriteString(point.Time.Format("15:04"))
		buf.WriteString(" ")
		ico, ok := icons[point.Weather]
		if !ok {
			buf.WriteString(point.Summary)
		} else {
			buf.WriteRune(rune(ico))
		}
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("%.1f℃", point.Temperature))
		buf.WriteString(fmt.Sprintf("/%.1f℃", point.ApparentTemperature))
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("%.0f%%", point.PrecipProbability*100))
		if point.Weather == weatherline.WeatherSnow {
			buf.WriteString(fmt.Sprintf("/%.0fcm", point.PrecipAccumulation))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

func createDaily(date time.Time, f *weatherline.ForecastResponse) string {
	if len(f.Daily.Data) == 0 {
		return ""
	}

	to := date.AddDate(0, 0, dateRange)
	var buf bytes.Buffer
	for _, point := range f.Daily.Data {
		d := truncHour(point.Time.Time)
		if !d.After(date) || !d.Before(to) && !d.Equal(to) {
			continue
		}

		buf.WriteString(point.Time.Format("01/02"))
		buf.WriteString(" ")
		ico, ok := icons[point.Weather]
		if !ok {
			buf.WriteString("??")
		} else {
			buf.WriteRune(rune(ico))
		}
		buf.WriteString("  ")
		buf.WriteString(fmt.Sprintf("%.0f%%", point.PrecipProbability*100))
		if point.Weather == weatherline.WeatherSnow {
			buf.WriteString(fmt.Sprintf("/%.0fcm", point.PrecipAccumulation))
		}
		buf.WriteString("\n")
		buf.WriteString("  ")
		buf.WriteString(fmt.Sprintf("%.1f℃", point.TemperatureHigh))
		buf.WriteString(fmt.Sprintf("/%.1f℃", point.ApparentTemperatureHigh))
		buf.WriteString(point.ApparentTemperatureHighTime.Format("(15:04)"))
		buf.WriteString("\n")
		buf.WriteString("  ")
		buf.WriteString(fmt.Sprintf("%.1f℃", point.TemperatureLow))
		buf.WriteString(fmt.Sprintf("/%.1f℃", point.ApparentTemperatureLow))
		buf.WriteString(point.ApparentTemperatureLowTime.Format("(15:04)"))
		buf.WriteString("\n")
		buf.WriteString("\n")
	}

	return buf.String()
}

func truncHour(t time.Time) time.Time {
	return t.Truncate(time.Hour).Add(time.Duration(-t.Hour()) * time.Hour)
}
