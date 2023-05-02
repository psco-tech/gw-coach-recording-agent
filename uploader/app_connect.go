package uploader

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type AppConnect struct {
	AgentToken 	string
	Host 		string
	client 		*http.Client
}

func NewAppConnect(agentToken string, host string) *AppConnect {
	return &AppConnect{
		client:      &http.Client{},
		AgentToken: agentToken,
		Host: host,
	}
}

func (a *AppConnect) makeRequest(path string, method string, body interface{}, respData interface{}) error {
	var reqBodyReader *strings.Reader
	if body != nil {
		reqBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBodyReader = strings.NewReader(string(reqBody))
	}else {
		reqBodyReader = strings.NewReader("")
	}

	req, err := http.NewRequest(method, a.Host +path, reqBodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-AGENT-AUTHORIZATION", "Bearer "+a.AgentToken)

	log.Default().Printf("Request INFO: ", req)

	resp, err := a.client.Do(req)
	if err != nil {
		log.Default().Printf("Request ERROR: ", err)
		return err
	}

	if resp.StatusCode >= 400 {
		var respError = *new(AppError)
		json.NewDecoder(resp.Body).Decode(&respError)
		return errors.New(respError.Error + " " + respError.Message)
	}

	defer resp.Body.Close()

	if respData != nil {
		log.Default().Printf("RESPONSE DATA : ", respData)
		err = json.NewDecoder(resp.Body).Decode(respData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *AppConnect) Info() (AgentInfo, error) {
	var info AgentInfo
	err := a.makeRequest("/u/agent/info", "GET", nil, &info)
	return info, err
}

func (a *AppConnect) GetTempUpload(upload TempUploadUrlRequest) (TempUploadUrlResponse, error){
	var resp TempUploadUrlResponse
	err := a.makeRequest("/u/agent/info", "GET", upload, &resp)
	return resp, err
}

func (a *AppConnect) FinalizeUpload(){

}
