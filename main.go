package main

import (
	yoBackClient "github.com/yoed/yoed-client-yo-back"
	clientInterface "github.com/yoed/yoed-client-interface"
	"net/http"
	"log"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strings"
	"time"
	"bytes"
	"math"
)

type TransilienYoedClient struct {
	yoBackClient.YoBackYoedClient
	config *TransilienYoedClientConfig
}

const Fmt = "15:04"
type configTime time.Time
func (ct *configTime) UnmarshalJSON(data []byte) error {
        b := bytes.NewBuffer(data)
        dec := json.NewDecoder(b)
        var s string
        if err := dec.Decode(&s); err != nil {
                return err
        }
        t, err := time.Parse(Fmt, s)
        if err != nil {
                return err
        }
        *ct = (configTime)(t)
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

type TransilienYoedClientConfig struct {
	FromStation string
	ToStation string
	Hour configTime
	Delta configDelta
}

type TransilienApiResponse struct {
    Data []*struct{
        TrainHour string
    }
}

func (c *TransilienYoedClient) loadConfig(configPath string) (*TransilienYoedClientConfig, error) {
	configJson, err := clientInterface.ReadConfig(configPath)

	if err != nil {
		return nil, err
	}

	config := &TransilienYoedClientConfig{}

	if err := json.Unmarshal(configJson, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *TransilienYoedClient) Handle(username string) {
	url := "http://transilien.ods.ocito.com/ods/transilien/iphone"
   	jsonDataIn := `[{"target":"/transilien/getNextTrains","map":{"codeArrivee":"PMP","codeDepart":"VMK"},"serial":"4"}]`
   	b := strings.NewReader(jsonDataIn)
	resp, err := http.Post(url, "application/json", b)

	log.Print("Requesting Transilien API from station %s to station %s", c.config.FromStation, c.config.ToStation)

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
        	c.YoBackYoedClient.Handle(username)
        }
	}
}

func (c *TransilienYoedClient) trainIsOnTime(data *TransilienApiResponse) bool {
	log.Printf("Parse data %v", data)

	hourPlusDelta := time.Time(c.config.Hour).Add(time.Duration(c.config.Delta))
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
		if time.Time(c.config.Hour).Equal(p) || math.Abs(float64(delta)) <= math.Abs(float64(c.config.Delta)) {
			log.Printf("On time!")
			return true
		}
	}

	return false
}

func NewTransilienYoedClient() (*TransilienYoedClient, error) {
	c := &TransilienYoedClient{}
	config, err := c.loadConfig("./config.json")

	if err != nil {
		panic(fmt.Sprintf("failed loading config: %s", err))
	}

	c.config = config
	baseClient, err := yoBackClient.NewYoBackYoedClient()

	if err != nil {
		return nil, err
	}
	c.YoBackYoedClient = *baseClient

	return c, nil
}

func main() {
	c, _ := NewTransilienYoedClient()

	clientInterface.Run(c)
}