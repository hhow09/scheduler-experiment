package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	pathRequestReport = "/request_report"
	pathGetReport     = "/get_report/%s" // /get_report/{ticketId}
)

const (
	programTimeout = 10 * time.Minute
	retryInterval  = 10 * time.Second
)

const timeFormat = "2006-01-02 15:04:05"

func main() {
	log("start report collector")
	reportApiUrl := os.Getenv("REPORT_API_URL")
	if reportApiUrl == "" {
		reportApiUrl = "http://localhost:8080"
		log(fmt.Sprintf("Defaulting to report api url %s", reportApiUrl))
	}
	ctx := context.Background()
	// the timeout of whole process
	ctx, cancel := context.WithTimeout(ctx, programTimeout)
	defer cancel()
	tickId, err := requestReport(ctx, reportApiUrl)
	if err != nil {
		log(fmt.Sprintf("error requesting report: %v", err))
		return
	}

	body, err := getReport(ctx, reportApiUrl, tickId)
	if err != nil {
		log(fmt.Sprintf("error getting report: %v", err))
		return
	}
	log(fmt.Sprintf("[report]\n%s", string(body)))
	// TODO to save the report to a file
}

func requestReport(ctx context.Context, reportApiUrl string) (string, error) {
	log("requesting report")
	// TODO set offset
	url := fmt.Sprintf("%s%s", reportApiUrl, pathRequestReport)
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("error requesting report: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	return string(body), nil
}

func getReport(ctx context.Context, reportApiUrl string, ticketId string) ([]byte, error) {
	log("getting report")
	url := fmt.Sprintf("%s%s", reportApiUrl, fmt.Sprintf(pathGetReport, ticketId))
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	for {
		resp, err := httpClient.Get(url)
		if err != nil {
			log(fmt.Sprintf("error downloading report: %v", err))
			return nil, fmt.Errorf("error downloading report: %v", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("[%s] error reading response body: %v\n", time.Now().Format(timeFormat), err)
			return nil, fmt.Errorf("error reading response body: %v", err)
		}
		if string(body) == "report is not ready" {
			log("report not ready, retrying...")
			time.Sleep(retryInterval)
			continue
		}
		return body, nil
	}
}

func log(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format(timeFormat), msg)
}
