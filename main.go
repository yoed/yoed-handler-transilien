package main

import (
	clientInterface "github.com/yoed/yoed-client-interface"
	"net/http"
	"log"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strings"
)

type TransilienYoedClient struct {
	clientInterface.BaseYoedClient
	config *TransilienYoedClientConfig
}

type TransilienYoedClientConfig struct {
	fromStation string
	toStation string
	hour string
	delta int
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

	log.Print("Requesting Transilien API from station %s to station %s", c.config.fromStation, c.config.toStation)

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
		var v []map[string]map[string]string
        json.Unmarshal(body, &v) 
        c.parseData(v[0]["data"])
        log.Printf("Transilien unmarshalled data %v", v)
	}
}

func (c *TransilienYoedClient) parseData(data map[string]string) {
	log.Printf("Parse data %v", data)
}

func NewTransilienYoedClient() (*TransilienYoedClient, error) {
	c := &TransilienYoedClient{}
	config, err := c.loadConfig("./config.json")

	if err != nil {
		panic(fmt.Sprintf("failed loading config: %s", err))
	}

	c.config = config
	baseClient, err := clientInterface.NewBaseYoedClient()

	if err != nil {
		return nil, err
	}
	c.BaseYoedClient = *baseClient

	return c, nil
}

func main() {
	c, _ := NewTransilienYoedClient()

	clientInterface.Run(c)
}