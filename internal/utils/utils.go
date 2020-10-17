package utils

import (
	"bindexer/internal/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// RPCHttpClient struct
type RPCHttpClient struct {
	hostName string
	username string
	password string
	port     int64
}

// NEW
// Returns new instance of RCPHttpClient
func (r *RPCHttpClient) New(hostName, username, password string, port int64) *RPCHttpClient {
	// Verify if connection exists
	_, err := http.Get(fmt.Sprintf("http://%s:%s", hostName, strconv.FormatInt(port, 10)))
	if err != nil {
		log.Fatal("Could not establish connection to bitcoin rpc")
	}
	return &RPCHttpClient{
		hostName: hostName,
		username: username,
		password: password,
		port:     port,
	}
}

// DoRequest Makes connection to the rpc host
// Returns Array Byte
func (r *RPCHttpClient) DoRequest(requestString string) ([]byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s", r.hostName, strconv.FormatInt(r.port, 10)), strings.NewReader(requestString))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(r.username, r.password)
	client := &http.Client{
		Timeout: time.Second * 80,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err

	}
	defer resp.Body.Close()
	serverOutPut, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err

	}
	return serverOutPut, err
}

// ConverRPCResponseToObject
// JSON RPC returns in this format
// {"result": {},"error" : ""} we are only interested in result object
// if it doesnt exists return blank object and error message
func ConvertRpcResponseToType(response []byte, object interface{}) error {
	var data map[string]interface{}
	err := json.Unmarshal(response, &data)
	if err != nil {
		return err
	}
	if data["error"] != nil {
		return err
	}

	if strings.HasPrefix(reflect.TypeOf(data["result"]).String(), "map") {
		jsonRes, err := json.Marshal(data["result"].(map[string]interface{}))
		if err != nil {
			return err
		}
		return json.Unmarshal(jsonRes, &object)

	}
	jsonRes, err := json.Marshal(data["result"])
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonRes, &object)
}

// LoadConfig loads the configuration from json file
// Returns a json representation of the config file
func LoadConfig(location string) (config models.Config, err error) {
	content, err := ioutil.ReadFile(location)
	if err != nil {
		return config, err
	}
	json.Unmarshal([]byte(content), &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
