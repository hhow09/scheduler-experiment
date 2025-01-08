package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
)

var (
	// simulate heavy load
	// if heavyLoad is true, the getReport API will return "not ready"
	heavyLoad = false
)

const timeFormat = "2006-01-02 15:04:05"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /request_report", requestReport)
	mux.HandleFunc("GET /get_report/{ticketId}", getReport)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log(fmt.Sprintf("Defaulting to port %s", port))
	}

	// simulate periodic heavy load
	go func() {
		for {
			time.Sleep(2 * time.Minute)
			heavyLoad = !heavyLoad
			fmt.Printf("[%s] currently server in heavy load: %t\n", time.Now().Format(time.RFC3339), heavyLoad)
		}
	}()

	// Start HTTP server.
	log(fmt.Sprintf("Listening on port %s", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log(fmt.Sprintf("error listening on port %s: %v", port, err))
	}
}

// API 1. requestReport accept requesting report and return a request id
// user can use the request id to get the report from getReport API
func requestReport(w http.ResponseWriter, r *http.Request) {
	log("request report")
	query := r.URL.Query()
	offset := query.Get("offset")
	if offset == "" {
		offset = "0"
	}

	// for simplicity, we use uuid + offset as ticketId
	// e.g. 123e4567-e89b-12d3-a456-426614174000_10000
	uuid, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ticketId := fmt.Sprintf("%s_%s", uuid.String(), offset)
	// return ticketid
	w.Write([]byte(ticketId))
}

// API 2. getReport accept request id and return the report
// if the report is not ready, it will return "not ready"
func getReport(w http.ResponseWriter, r *http.Request) {
	// get ticketId from path
	ticketId := r.PathValue("ticketId")
	log(fmt.Sprintf("get report, ticketId: %s", ticketId))
	// parse ticketId to get uuid and offset
	_, offset, err := parseTicketId(ticketId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if heavyLoad {
		w.Write([]byte("report is not ready"))
		return
	}

	// generate some csv data
	reportCSV, err := getReportFromOffset(offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return csv data
	fileName := fmt.Sprintf("%d.csv", offset)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	w.Write(reportCSV)
}

func parseTicketId(ticketId string) (string, int, error) {
	parts := strings.Split(ticketId, "_")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid ticketId format")
	}
	offsetInt, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid offset format")
	}
	return parts[0], offsetInt, nil
}

func getReportFromOffset(offset int) ([]byte, error) {
	time.Sleep(1 * time.Second)
	type row struct {
		Timestamp int    `csv:"timestamp"`
		Data      string `csv:"data"`
	}
	rows := []row{}
	for i := 0; i < 5; i++ {
		rows = append(rows, row{
			Timestamp: offset + i,
			Data:      fmt.Sprintf("data-%d", i),
		})
	}
	// generate some csv data
	w := bytes.NewBuffer(nil)
	gocsv.Marshal(rows, w)
	return w.Bytes(), nil
}

func log(msg string) {
	fmt.Printf("[%s] %s. heavyLoad: %t\n", time.Now().Format(timeFormat), msg, heavyLoad)
}
