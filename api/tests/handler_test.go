/*
 * Copyright (C) 2018 eeonevision
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */
package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/eeonevision/anychaindb/api/handler"
)

var host = flag.String("host", "localhost", "machine host")
var apiMode = flag.String("apimode", "sync", "api requests processing mode")
var apiPort = flag.String("apiport", "26659", "api port")
var rpcPort = flag.String("rpcport", "26657", "rpc port")
var update = flag.Bool("update", false, "update .golden files")

var errNotFound = errors.New("404")

func doPOSTRequest(endpoint, url, id, privKey string, data []byte) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * 30}
	req, _ := http.NewRequest("POST", "http://"+url+endpoint+"?mode="+*apiMode, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if id != "" && privKey != "" {
		req.SetBasicAuth(id, privKey)
	}
	// Send request to Anychaindb
	respRaw, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer respRaw.Body.Close()
	contents, err := ioutil.ReadAll(respRaw.Body)
	if err != nil {
		return nil, err
	}
	// Check response status code
	if respRaw.StatusCode != http.StatusAccepted {
		return contents, errors.New(strconv.Itoa(respRaw.StatusCode))
	}
	return contents, nil
}

func doGETRequest(endpoint, url, id, privKey string) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * 30}
	req, _ := http.NewRequest("GET", "http://"+url+endpoint, nil)
	if id != "" && privKey != "" {
		req.SetBasicAuth(id, privKey)
	}
	respRaw, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer respRaw.Body.Close()
	contents, err := ioutil.ReadAll(respRaw.Body)
	if err != nil {
		return nil, err
	}
	// Check response status code
	if respRaw.StatusCode != http.StatusOK {
		return contents, errors.New(strconv.Itoa(respRaw.StatusCode))
	}
	return contents, nil
}

func TestCreateAccount(t *testing.T) {
	// Generate transaction request
	endpoint := fmt.Sprintf("/v1/accounts")
	url := *host + ":" + *apiPort
	data, _ := json.Marshal(handler.Request{})
	contents, err := doPOSTRequest(endpoint, url, "", "", data)
	if err != nil {
		t.Errorf("error in sending POST request: %s", contents)
		return
	}
	// Check data in results
	resp := handler.Result{}
	err = json.Unmarshal(contents, &resp)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	account := resp.Data.(map[string]interface{})
	if account["ID"] == "" {
		t.Errorf("account has no ID: %s", account)
		return
	}
	// Write account data to account.golden file
	if *update {
		// Create folder if not exists
		_ = os.Mkdir("artifacts", os.ModePerm)
		// Marshal account data to JSON string
		res, err := json.Marshal(resp.Data)
		if err != nil {
			t.Errorf("%s", err)
			return
		}
		// Write account id
		err = ioutil.WriteFile("artifacts/account.golden", res, 0644)
		if err != nil {
			t.Errorf("%s", err)
			return
		}
		// Wait for transaction approve
		time.Sleep(time.Second * 5)
	}
	t.Logf("added account: %v", account)
}

func TestGetAccount(t *testing.T) {
	// Generate transaction request
	endpoint := fmt.Sprintf("/v1/accounts")
	url := *host + ":" + *apiPort
	// Read account.golden file
	file, err := ioutil.ReadFile("artifacts/account.golden")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Parse account.golden file
	acc1 := handler.Account{}
	err = json.Unmarshal(file, &acc1)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if len(acc1.ID) == 0 {
		t.Errorf("account.golden is empty: %s", acc1)
		return
	}
	endpoint = endpoint + "/" + acc1.ID
	// Find account in Anychaindb server
	contents, err := doGETRequest(endpoint, url, "", "")
	if err != nil {
		t.Errorf("error in sending GET request: %s", contents)
		return
	}
	resp := handler.Result{}
	err = json.Unmarshal(contents, &resp)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Compare accounts
	acc2 := resp.Data.(map[string]interface{})
	if acc2["_id"] != acc1.ID || acc2["public_key"] != acc1.Pub {
		t.Errorf("accounts are not equal. Expected: %v, Output: %v", acc1, acc2)
		return
	}
	t.Logf("got account: %v", acc2)
}

func TestDetectNotFoundAccount(t *testing.T) {
	// Generate transaction request
	endpoint := fmt.Sprintf("/v1/accounts")
	url := *host + ":" + *apiPort
	endpoint = endpoint + "/34"
	// Find account in Anychaindb server
	_, err := doGETRequest(endpoint, url, "", "")
	if err.Error() != errNotFound.Error() {
		t.Errorf("error in getting wrong account: %s", err)
		return
	}
	t.Logf("returned correct status code for not found account: %s", err)
}

type conversion struct {
	OfferID  int `json:"offer_id"`
	StreamID int `json:"stream_id"`
}

func TestCreatePayload(t *testing.T) {
	// Generate transaction request
	endpoint := fmt.Sprintf("/v1/payloads")
	url := *host + ":" + *apiPort
	// Read account.golden file
	file, err := ioutil.ReadFile("artifacts/account.golden")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Parse account.golden file
	acc1 := handler.Account{}
	err = json.Unmarshal(file, &acc1)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if len(acc1.ID) == 0 {
		t.Errorf("account.golden is empty: %s", acc1)
		return
	}
	// Send request
	data, _ := json.Marshal(handler.Request{
		AccountID: acc1.ID,
		PrivKey:   acc1.Priv,
		PubKey:    acc1.Pub,
		Data: handler.Payload{
			PublicData: "test_public_data",
			PrivateData: []*handler.PrivateData{
				&handler.PrivateData{
					ReceiverAccountID: acc1.ID,
					Data:              conversion{OfferID: 15, StreamID: 1},
				},
			},
		}})
	contents, err := doPOSTRequest(endpoint, url, "", "", data)
	if err != nil {
		t.Errorf("error in sending POST request: %s", contents)
		return
	}
	// Check data in results
	resp := handler.Result{}
	err = json.Unmarshal(contents, &resp)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	cnv := resp.Data.(map[string]interface{})
	if cnv["_id"] == "" {
		t.Errorf("payload has no id: %s", cnv)
		return
	}
	// Write payload to payload.golden file
	if *update {
		// Create folder if not exists
		_ = os.Mkdir("artifacts", os.ModePerm)
		// Marshal payload to JSON string
		res, err := json.Marshal(resp.Data)
		if err != nil {
			t.Errorf("%s", err)
			return
		}
		// Write payload id
		err = ioutil.WriteFile("artifacts/payload.golden", res, 0644)
		if err != nil {
			t.Errorf("%s", err)
			return
		}
		// Wait for transaction approve
		time.Sleep(time.Second * 5)
	}
	t.Logf("added payload: %v", string(data))
}

func TestGetPayload(t *testing.T) {
	// Generate transaction request
	endpoint := fmt.Sprintf("/v1/payloads")
	url := *host + ":" + *apiPort
	// Read payload.golden file
	file, err := ioutil.ReadFile("artifacts/payload.golden")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Parse payload.golden file
	cnv1 := handler.Payload{}
	err = json.Unmarshal(file, &cnv1)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if len(cnv1.ID) == 0 {
		t.Errorf("payload.golden is empty: %v", cnv1)
		return
	}
	endpoint = endpoint + "/" + cnv1.ID
	// Find payload in Anychaindb server
	contents, err := doGETRequest(endpoint, url, "", "")
	if err != nil {
		t.Errorf("error in sending GET request: %s", contents)
		return
	}
	resp := handler.Result{}
	err = json.Unmarshal(contents, &resp)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Compare payloads
	cnv2 := resp.Data.(map[string]interface{})
	if cnv2["_id"] != cnv1.ID {
		t.Errorf("payloads are not equal. Expected: %v, Output: %v", cnv1.ID, cnv2["_id"])
		return
	}
	t.Logf("got payload: %v", cnv2)
}

func TestGetDecryptedPayload(t *testing.T) {
	// Generate transaction request
	endpoint := fmt.Sprintf("/v1/payloads")
	url := *host + ":" + *apiPort
	// Read payload.golden file
	plFile, err := ioutil.ReadFile("artifacts/payload.golden")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Parse payload.golden file
	cnv1 := handler.Payload{}
	err = json.Unmarshal(plFile, &cnv1)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if len(cnv1.ID) == 0 {
		t.Errorf("payload.golden is empty: %v", cnv1)
		return
	}
	// Read account.golden file
	accFile, err := ioutil.ReadFile("artifacts/account.golden")
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Parse account.golden file
	acc1 := handler.Account{}
	err = json.Unmarshal(accFile, &acc1)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if len(acc1.ID) == 0 {
		t.Errorf("account.golden is empty: %s", acc1)
		return
	}
	endpoint = endpoint + "/" + cnv1.ID
	// Find payload in Anychaindb server
	contents, err := doGETRequest(endpoint, url, acc1.ID, acc1.Priv)
	if err != nil {
		t.Errorf("error in sending GET request: %s", contents)
		return
	}
	resp := handler.Result{}
	err = json.Unmarshal(contents, &resp)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	// Compare payloads
	t.Log(string(contents))
	cnv2 := resp.Data.(map[string]interface{})
	if cnv2["_id"] != cnv1.ID {
		t.Errorf("payloads are not equal. Expected: %v, Output: %v", cnv1.ID, cnv2["_id"])
		return
	}
	// TODO: Compare payload data
	t.Logf("got payload: %v", cnv2)
}

func TestDetectNotFoundPayload(t *testing.T) {
	// Generate transaction request
	endpoint := fmt.Sprintf("/v1/payloads")
	url := *host + ":" + *apiPort
	endpoint = endpoint + "/34"
	// Find payload in Anychaindb server
	_, err := doGETRequest(endpoint, url, "", "")
	if err.Error() != errNotFound.Error() {
		t.Errorf("error in getting wrong payload: %s", err)
		return
	}
	t.Logf("returned correct status code for not found payload: %s", err)
}
