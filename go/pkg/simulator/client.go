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
	//logrus.SetLevel(logrus.InfoLevel)

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
	return IssueRequest(c.RestyClient, "POST", "scan", scan, nil, nil)
}

func (c *Client) FetchScan(scanId string) (*ScanResults, error) {
	scan := &ScanResults{}
	_, err := IssueRequest(c.RestyClient, "GET", "scan", nil, map[string]string{"scan-id": scanId}, scan)
	return scan, err
}

func (c *Client) Start() {
	for w := 0; w < 10; w++ {
		go func(workerId int) {
			workerIdString := fmt.Sprintf("%d", workerId)
			for i := 0; ; i++ {
				data := rand.String(40_000)

				RecordEventValue("issuing request", workerIdString, float64(i))
				logrus.Infof("issuing request: %d, %d, %s", workerId, i, data[:15])

				resp, err := c.StartScan(&StartScan{
					Data: fmt.Sprintf("%d-%d-%s", workerId, i, data),
				})
				logrus.Infof("response from request %d, %d: %s", workerId, i, resp)
				RecordEvent("start error", workerIdString, err)
				if err != nil {
					logrus.Errorf("unable to start scan: %+v", err)
				}

				scan, err := c.FetchScan(fmt.Sprintf("%d-%d", workerId, i))
				RecordEvent("fetch scan", workerIdString, err)
				if err != nil {
					logrus.Errorf("unable to fetch scan: %+v", err)
				} else {
					scanString := scan.Data
					logrus.Infof("scan string: %s of %d", scanString[:15], len(scanString))
				}
				time.Sleep(1 * time.Second)
			}
		}(w)
	}
}

func IssueRequest(restyClient *resty.Client, verb string, path string, body interface{}, params map[string]string, result interface{}) (string, error) {
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
	request = request.SetQueryParams(params)

	urlPath := fmt.Sprintf("%s/%s", restyClient.BaseURL, path)
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
	logrus.Tracef("response body: %s, %+v", respBody, result)

	if !resp.IsSuccess() {
		return respBody, errors.Errorf("bad status code for %s to path %s: %d, response %s", verb, path, statusCode, respBody)
	}
	return respBody, nil
}
