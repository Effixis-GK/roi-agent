package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// wasIntervalTransmitted checks if an interval was already successfully transmitted
func (ds *DataSender) wasIntervalTransmitted(startTime, endTime time.Time) bool {
	logs, err := ds.loadTransmissionLogs()
	if err != nil {
		return false
	}

	for _, logEntry := range logs {
		if logEntry.Success && 
		   logEntry.StartTime.Equal(startTime) && 
		   logEntry.EndTime.Equal(endTime) {
			return true
		}
	}

	return false
}

// loadTransmissionLogs loads transmission log history
func (ds *DataSender) loadTransmissionLogs() ([]TransmissionLog, error) {
	data, err := ioutil.ReadFile(ds.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []TransmissionLog{}, nil
		}
		return nil, err
	}

	var logs []TransmissionLog
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// logTransmissionResult logs the result of a transmission attempt
func (ds *DataSender) logTransmissionResult(startTime, endTime time.Time, success bool, err error, retryCount, payloadSize int) {
	logs, _ := ds.loadTransmissionLogs()

	logEntry := TransmissionLog{
		StartTime:   startTime,
		EndTime:     endTime,
		Timestamp:   time.Now(),
		Success:     success,
		RetryCount:  retryCount,
		PayloadSize: payloadSize,
	}

	if err != nil {
		logEntry.Error = err.Error()
	}

	logs = append(logs, logEntry)

	// Keep only last 100 log entries
	if len(logs) > 100 {
		logs = logs[len(logs)-100:]
	}

	data, jsonErr := json.MarshalIndent(logs, "", "  ")
	if jsonErr != nil {
		log.Printf("Error marshaling transmission logs: %v", jsonErr)
		return
	}

	if writeErr := ioutil.WriteFile(ds.logPath, data, 0644); writeErr != nil {
		log.Printf("Error saving transmission logs: %v", writeErr)
	}
}
