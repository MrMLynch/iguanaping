package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type NtxCount []struct {
	NnName string `json:"nn_name"`
	Total  int    `json:"total"`
	Period string `json:"period"`
}

func NtxCountFunc() NtxCount {
	url := "https://notarystats.info/api/testnet.php?period=24h"

	ntxClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "nn-testnet")

	res, getErr := ntxClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	ops := NtxCount{}
	jsonErr := json.Unmarshal(body, &ops)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return ops
}
