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
)

type Handler struct {
	yoBackHandler.Handler
	Config *Config
}

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
	yoBackHandler.Config
	FromStation string `json:"from_station"`
	ToStation string `json:"to_station"`
	Interval configInterval
	Delta configDelta
}

type TransilienApiResponse struct {
    Data []*struct{
        TrainHour string
    }
}

func (c *Handler) Handle(username, handle string) {
	url := "http://transilien.ods.ocito.com/ods/transilien/iphone"
   	jsonDataIn := `[{"target":"/transilien/getNextTrains","map":{"codeArrivee":"`+c.Config.ToStation+`","codeDepart":"`+c.Config.FromStation+`"},"serial":"4"}]`
   	b := strings.NewReader(jsonDataIn)
	resp, err := http.Post(url, "application/json", b)

	log.Printf("Requesting Transilien API from station %s to station %s", c.Config.FromStation, c.Config.ToStation)

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
		// For testing purposes
		// hourPlusInterval := time.Now().Add(time.Duration(c.Config.Interval)).Round(time.Duration(time.Minute))
		// mockBody := []byte("[{\"data\":[{\"trainDock\":null,\"trainHour\":\""+hourPlusInterval.Format("02/01/2006 15:04")+"\",\"trainLane\":\"2B\",\"trainMention\":null,\"trainMissionCode\":\"PORO\",\"trainNumber\":\"165484\",\"trainTerminus\":\"PMP\",\"type\":\"R\"}]}]")
        json.Unmarshal(body, &v) 
        log.Printf("Transilien unmarshalled data %v", v)
        if c.trainIsOnTime(v[0]) {
        	c.Handler.Handle(username, handle)
        }
	}
}

func (c *Handler) trainIsOnTime(data *TransilienApiResponse) bool {
	log.Printf("Parse data %v", data)

	intervalDuration := time.Duration(c.Config.Interval)
	frenchLocation, _ := time.LoadLocation("Europe/Paris") 

	hourPlusInterval := time.Now().Add(intervalDuration).Round(time.Duration(time.Minute))
	hourPlusDelta := hourPlusInterval.Add(time.Duration(c.Config.Delta))
	for _, aData := range data.Data {
		log.Printf("Data: %v", aData.TrainHour)
		p, err := time.ParseInLocation("02/01/2006 15:04", string(aData.TrainHour), frenchLocation) // API's time have no timezone
		if err != nil {
			log.Printf("Error parsing train time: %s", err)
			continue
		}

		if (hourPlusInterval.Equal(p) || hourPlusInterval.Before(p)) && (hourPlusDelta.Equal(p) || hourPlusDelta.After(p)) {
			log.Printf("On time!")
			return true
		}
	}

	log.Printf("No train in %s", intervalDuration)
	return false
}

func New() *Handler {

	c := &Handler{}

	if err := httpInterface.LoadConfig("./config.json", &c.Config); err != nil {
		log.Fatalf("failed loading config: %s", err)
	}

	c.Handler.Config = &c.Config.Config

	return c
}