package listerwatcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"minik8s/pkg/apiserver/app"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/messaging"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type WatchHandler func(res etcdstore.WatchRes)
type CancelFunc func()

type ListerWatcher struct {
	subscriber *messaging.Subscriber
	rootURL    string
}

func NewListerWatcher(c *Config) (*ListerWatcher, error) {
	s, err := messaging.NewSubscriber(c.QueueConfig)
	if err != nil {
		return nil, err
	}
	ls := &ListerWatcher{
		subscriber: s,
		rootURL:    fmt.Sprintf("http://%s:%d", c.Host, c.HttpPort),
	}
	return ls, nil
}

func (ls *ListerWatcher) List(key string) ([]etcdstore.ListRes, error) {
	resourceURL := ls.rootURL + key
	request, err := http.NewRequest("GET", resourceURL, nil)
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

// Watch should never return until stopChannel is closed
func (ls *ListerWatcher) Watch(key string, handler WatchHandler, stopChannel <-chan struct{}) error {
	// request the server to publish
	resourceURL := ls.rootURL + key
	request, err := http.NewRequest("POST", resourceURL, nil)
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
	reader := response.Body
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	var t app.Ticket
	err = json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	defer func() {
		formData := url.Values{}
		klog.Infof("Closing ticket %d\n", t.T)
		formData.Add("ticket", strconv.FormatUint(t.T, 10))
		response, err := http.DefaultClient.Post(resourceURL, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
		if err != nil {
			klog.Errorf("Error [%s] closing the watch channel with ticket %d\n", err.Error(), t.T)
		} else if response.StatusCode != http.StatusOK {
			klog.Errorf("Status Code %d !\n", response.StatusCode)
		}
	}()

	// received response from server
	// begin to subscribe
	stop := make(chan struct{})
	amqpHandler := func(d amqp.Delivery) {
		var res etcdstore.WatchRes
		err := json.Unmarshal(d.Body, &res)
		if err != nil {
			klog.Errorf("Error [%s] unmarshalling data from amqp channel\n", err.Error())
			return
		}
		handler(res)
	}
	err = ls.subscriber.Subscribe(key, amqpHandler, stop)
	if err != nil {
		return err
	}

	defer func() {
		ls.subscriber.Unsubscribe(stop)
	}()

	<-stopChannel
	return nil
}

// WatchNonBlocking is a non-blocking version of watch mechanism.
//
// It will return with am unwatch function and a nil err if watching etcd normally.
//
// Remember to call the unwatch function to release resources!
//
// You will get a nil unwatch function and a non-nil err when something unexpected happens.
//
// Attention : Different from the Watch above, no matter what happens, this will always return immediately!
func (ls *ListerWatcher) WatchNonBlocking(key string, handler WatchHandler) (CancelFunc, error) {
	// request the server to publish
	resourceURL := ls.rootURL + key
	request, err := http.NewRequest("POST", resourceURL, nil)
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
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var t app.Ticket
	err = json.Unmarshal(data, &t)
	if err != nil {
		return nil, err
	}

	withdrawTicket := func() {
		formData := url.Values{}
		klog.Infof("Closing ticket %d\n", t.T)
		formData.Add("ticket", strconv.FormatUint(t.T, 10))
		response, err := http.DefaultClient.Post(resourceURL, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
		if err != nil {
			klog.Errorf("Error [%s] closing the watch channel with ticket %d\n", err.Error(), t.T)
		} else if response.StatusCode != http.StatusOK {
			klog.Errorf("Status Code %d !\n", response.StatusCode)
		}
	}

	// received response from server
	// begin to subscribe
	stop := make(chan struct{})
	amqpHandler := func(d amqp.Delivery) {
		var res etcdstore.WatchRes
		err := json.Unmarshal(d.Body, &res)
		if err != nil {
			klog.Errorf("Error [%s] unmarshalling data from amqp channel\n", err.Error())
			return
		}
		handler(res)
	}
	err = ls.subscriber.Subscribe(key, amqpHandler, stop)
	if err != nil {
		withdrawTicket()
		return nil, err
	}

	return func() {
		withdrawTicket()
		ls.subscriber.Unsubscribe(stop)
	}, nil
}
