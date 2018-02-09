package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Configuration struct {
	Debug          bool   `json:"debug"`
	KeyID          string `json:"key_id"`
	CountryOrState string `json:"country_or_state"`
	Location       string `json:"location"`
}

type Wunderground struct {
	HourlyForecasts []HourlyForecast `json:"hourly_forecast"`
}

type Time struct {
	Hour  int    `json:"hour,string"`
	Day   int    `json:"mday,string"`
	Month string `json:"month_name_abbrev"`
}

type TempForecast struct {
	Temp int `json:"metric,string"`
}

type FeelsLikeForecast struct {
	FeelsLike int `json:"metric,string"`
}

type WindSpeedForecast struct {
	WindSpeed int `json:"metric,string"`
}

type WindDirForecast struct {
	WindDir string `json:"dir"`
}

type HourlyForecast struct {
	Time              Time              `json:"FCTTIME"`
	Humidity          int               `json:"humidity,string"`
	Condition         string            `json:"condition"`
	TempForecast      TempForecast      `json:"temp"`
	FeelsLikeForecast FeelsLikeForecast `json:"feelslike"`
	PercpForecast     int               `json:"pop,string"`
	WindSpeedForecast WindSpeedForecast `json:"wspd"`
	WindDirForecast   WindDirForecast   `json:"wdir"`
}

func GetHourlyForecasts(key, country, loc string) []HourlyForecast {
	url := fmt.Sprintf("https://api.wunderground.com/api/%s/hourly/q/%s/%s.json", key, country, loc)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var w Wunderground
	if err := json.Unmarshal(body, &w); err != nil {
		log.Fatal(err)
	}
	return w.HourlyForecasts
}

func FilterForecasts(forecasts []HourlyForecast, day, startHour, endHour int) []HourlyForecast {
	var f []HourlyForecast
	for _, h := range forecasts {
		if h.Time.Day == day && h.Time.Hour >= startHour && h.Time.Hour <= endHour {
			f = append(f, h)
		}
	}
	return f
}

func main() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	var config Configuration
	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatal(err)
	}

	f := GetHourlyForecasts(config.KeyID, config.CountryOrState, config.Location)

	// Show today's (0:00-11:00) or tomorrow's (12:00-0:00) forecast.
	now := time.Now().UTC()
	var day int
	var month time.Month
	if now.Hour() < 11 {
		_, month, day = now.Date()
	} else {
		_, month, day = now.Add(24 * time.Hour).Date()
	}
	filter := FilterForecasts(f, day, 6, 12)

	date := fmt.Sprintf("%d %s", day, month.String())
	router := gin.Default()
	router.LoadHTMLFiles("forecast.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "forecast.tmpl", gin.H{"day": date, "forecasts": filter})
	})
	router.Run(":" + os.Getenv("PORT"))
}
