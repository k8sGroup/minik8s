package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"minik8s/pkg/etcdstore"
	"net/http"
	url2 "net/url"
)

func Get(url string) ([]etcdstore.ListRes, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("StatusCode not 200")
	}
	reader := response.Body
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var resList []etcdstore.ListRes
	err = json.Unmarshal(data, &resList)
	if err != nil {
		return nil, err
	}
	return resList, nil
}

func GetWithParams(url string, params map[string]string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	values := url2.Values{}
	for key, val := range params {
		values.Add(key, val)
	}
	request.URL.RawQuery = values.Encode()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("StatusCode not 200")
	}
	reader := response.Body
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Del(url string) error {
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("StatusCode not 200")
	}
	return nil
}

func Put(url string, obj any) error {
	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(payload)
	request, err := http.NewRequest("PUT", url, reader)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("StatusCode not 200")
	}
	return nil
}
