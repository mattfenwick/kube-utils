package simulator

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/rand"
	"time"
)

func RunClient(serverAddress string) {
	client := NewClient(serverAddress)
	client.Start()

	stop := make(chan struct{})
	<-stop
}

type Client struct {
	ServerAddress string
	RestyClient   *resty.Client
}

func NewClient(server string) *Client {
	restyClient := resty.New()
	restyClient.SetBaseURL(server)
	return &Client{
		ServerAddress: server,
		RestyClient:   restyClient,
	}
}

func (c *Client) StartScan(scan *StartScan) (string, error) {
	return IssueRequest(c.RestyClient, "POST", "scan", scan, nil)
}

func (c *Client) Start() {
	for w := 0; w < 10; w++ {
		go func(workerId int) {
			for i := 0; ; i++ {
				logrus.Infof("issuing request: %d, %d", workerId, i)
				resp, err := c.StartScan(&StartScan{
					Data: fmt.Sprintf("%d-%d-%s", workerId, i, rand.String(40_000)),
				})
				logrus.Infof("response to request %d, %d: %s", workerId, i, resp)
				if err != nil {
					logrus.Errorf("error? %+v", err)
				}
				time.Sleep(1 * time.Second)
			}
		}(w)
	}
}

func IssueRequest(restyClient *resty.Client, verb string, path string, body interface{}, result interface{}) (string, error) {
	var err error
	request := restyClient.R()
	if body != nil {
		reqBody, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			return "", errors.Wrapf(err, "unable to marshal json")
		}
		logrus.Tracef("request body: %s", string(reqBody))
		request = request.SetBody(body)
	}
	if result != nil {
		request = request.SetResult(result)
	}

	urlPath := fmt.Sprintf("%s/%s", restyClient.HostURL, path)
	logrus.Debugf("issuing %s to %s", verb, urlPath)

	var resp *resty.Response
	switch verb {
	case "GET":
		resp, err = request.Get(path)
	case "POST":
		resp, err = request.Post(path)
	case "PUT":
		resp, err = request.Put(path)
	case "DELETE":
		resp, err = request.Delete(path)
	default:
		return "", errors.Errorf("unrecognized http verb %s to %s", verb, path)
	}
	if err != nil {
		return "", errors.Wrapf(err, "unable to issue %s to %s", verb, path)
	}

	respBody, statusCode := resp.String(), resp.StatusCode()
	logrus.Debugf("response code %d from %s to %s", statusCode, verb, urlPath)
	logrus.Tracef("response body: %s", respBody)

	if !resp.IsSuccess() {
		return respBody, errors.Errorf("bad status code for %s to path %s: %d, response %s", verb, path, statusCode, respBody)
	}
	return respBody, nil
}
