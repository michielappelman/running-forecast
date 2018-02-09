package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/LoganK/go-wunderground"
	"github.com/gin-gonic/gin"
)

type Configuration struct {
	Debug          bool   `json:"debug"`
	KeyID          string `json:"key_id"`
	CountryOrState string `json:"country_or_state"`
	Location       string `json:"location"`
}

func GetHourlyForecasts(key, country, city string) *wunderground.HourlyForecast {
	q := wunderground.Query{Country: country, City: city}
	w := wunderground.NewService(key)
	a, err := w.Request([]string{"hourly"}, &q)
	if err != nil {
		log.Fatal(err)
	}

	return a.HourlyForecast
}

func FilterForecasts(forecasts *wunderground.HourlyForecast, day, startHour, endHour int) wunderground.HourlyForecast {
	var f wunderground.HourlyForecast
	for _, hourly := range *forecasts {
		d, _ := strconv.Atoi(hourly.FCTTIME.Mday)
		h, _ := strconv.Atoi(hourly.FCTTIME.Hour)
		if d == day && h >= startHour && h <= endHour {
			f = append(f, hourly)
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
