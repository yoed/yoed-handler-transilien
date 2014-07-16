package handler

import (
	yoBackHandler "github.com/yoed/yoed-handler-yo-back/handler"
	httpInterface "github.com/yoed/yoed-http-interface"

	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"strings"
	"time"
	"bytes"
	"math"
)

type Handler struct {
	yoBackHandler.Handler
	Config *Config
}

const Fmt = "15:04"
type configInterval time.Duration
func (ct *configInterval) UnmarshalJSON(data []byte) error {
    b := bytes.NewBuffer(data)
    dec := json.NewDecoder(b)
    var s string
    if err := dec.Decode(&s); err != nil {
            return err
    }
    t, err := time.ParseDuration(s)
    if err != nil {
       return err
    }
    *ct = (configInterval)(t)
    return nil
}

type configDelta time.Duration
func (cd *configDelta) UnmarshalJSON(data []byte) error {
    b := bytes.NewBuffer(data)
    dec := json.NewDecoder(b)
    var s string
    if err := dec.Decode(&s); err != nil {
            return err
    }
    t, err := time.ParseDuration(s)
    if err != nil {
            return err
    }
    *cd = (configDelta)(t)
    return nil
}

type Config struct {
	httpInterface.Config
	FromStation string
	ToStation string
	Interval configInterval
	Delta configDelta
}

type TransilienApiResponse struct {
    Data []*struct{
        TrainHour string
    }
}

func (c *Handler) Handle(username string) {
	url := "http://transilien.ods.ocito.com/ods/transilien/iphone"
   	jsonDataIn := `[{"target":"/transilien/getNextTrains","map":{"codeArrivee":"`+c.Config.ToStation+`","codeDepart":"`+c.Config.FromStation+`"},"serial":"4"}]`
   	b := strings.NewReader(jsonDataIn)
	resp, err := http.Post(url, "application/json", b)

	log.Print("Requesting Transilien API from station %s to station %s", c.Config.FromStation, c.Config.ToStation)

	if err != nil {
		log.Printf("Transilien API error: %s", err)
		return
	}

	defer resp.Body.Close()

	log.Printf("Transilien API response status: %s", resp.Status)

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Printf("Transilien API response body error: %s", err)
	} else {
		log.Printf("Transilien API response body: %s", string(body))
		var v []*TransilienApiResponse
        json.Unmarshal(body, &v) 
        log.Printf("Transilien unmarshalled data %v", v)
        if c.trainIsOnTime(v[0]) {
        	c.Handler.Handle(username)
        }
	}
}

func (c *Handler) trainIsOnTime(data *TransilienApiResponse) bool {
	log.Printf("Parse data %v", data)

	hourPlusInterval := time.Now().Add(time.Duration(c.Config.Interval))
	hourPlusDelta := time.Now().Add(time.Duration(c.Config.Delta))
	for _, aData := range data.Data {
		log.Printf("Data: %v", aData.TrainHour)
		p, err := time.Parse("02/01/2006 15:04", string(aData.TrainHour))
		if err != nil {
			log.Printf("Error parsing train time: %s", err)
			continue
		}

		trainHourParsed, err := time.Parse(Fmt, p.Format(Fmt))
		if err != nil {
			log.Printf("Error parsing train hour: %s", err)
			continue
		}
		delta := trainHourParsed.Sub(hourPlusDelta)
		if hourPlusInterval.Equal(p) || math.Abs(float64(delta)) <= math.Abs(float64(c.Config.Delta)) {
			log.Printf("On time!")
			return true
		}
	}

	return false
}

func New() *Handler {

	c := &Handler{}

	if err := httpInterface.LoadConfig("./config.json", &c.Config); err != nil {
		log.Fatalf("failed loading config: %s", err)
	}

	return c
}