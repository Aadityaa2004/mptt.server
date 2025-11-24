package weather

// ForecastMain represents the main weather data for forecast items
type ForecastMain struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	SeaLevel  int     `json:"sea_level"`
	GrndLevel int     `json:"grnd_level"`
	Humidity  int     `json:"humidity"`
	TempKf    float64 `json:"temp_kf"`
}

// ForecastSys represents system data for forecast items
type ForecastSys struct {
	Pod string `json:"pod"` // "d" for day, "n" for night
}

// Rain represents rain data
type Rain struct {
	ThreeH float64 `json:"3h"`
}

// Snow represents snow data
type Snow struct {
	ThreeH float64 `json:"3h"`
}

// ForecastItem represents a single forecast entry in the list
type ForecastItem struct {
	Dt         int64         `json:"dt"`
	Main       ForecastMain  `json:"main"`
	Weather    []WeatherItem `json:"weather"`
	Clouds     Clouds        `json:"clouds"`
	Wind       Wind          `json:"wind"`
	Visibility int           `json:"visibility"`
	Pop        float64       `json:"pop"` // Probability of precipitation
	Sys        ForecastSys   `json:"sys"`
	DtTxt      string        `json:"dt_txt"`
	Rain       *Rain         `json:"rain,omitempty"` // Optional, only present when raining
	Snow       *Snow         `json:"snow,omitempty"` // Optional, only present when snowing
}

// City represents city information in the forecast response
type City struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Coord      Coord  `json:"coord"`
	Country    string `json:"country"`
	Population int    `json:"population"`
	Timezone   int    `json:"timezone"`
	Sunrise    int64  `json:"sunrise"`
	Sunset     int64  `json:"sunset"`
}

// Forecast represents the complete forecast response
type Forecast struct {
	Cod     string         `json:"cod"`
	Message int            `json:"message"`
	Cnt     int            `json:"cnt"`
	List    []ForecastItem `json:"list"`
	City    City           `json:"city"`
}
