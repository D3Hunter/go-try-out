package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/docker/go-units"
)

type Time struct {
	time.Time
}

const logTimeFormat = "2006/01/02 15:04:05.000 -07:00"

func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	// TODO(https://go.dev/issue/47353): Properly unescape a JSON string.
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("Time.UnmarshalJSON: input is not a JSON string")
	}
	data = data[len(`"`) : len(data)-len(`"`)]
	var err error
	t.Time, err = time.Parse(logTimeFormat, string(data))
	return err
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(data []byte) (err error) {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	d.Duration, err = time.ParseDuration(s)
	return err
}

type logHeader struct {
	Time    Time   `json:"time"`
	Message string `json:"message"`
}

type costLog struct {
	logHeader `json:",inline"`
	Cost      Duration `json:"cost"`
	Call      string   `json:"call"`
	JobID     int64    `json:"job_id"`
	JobState  string   `json:"job_state"`
	SQL       string   `json:"sql"`
	Version   int64    `json:"version"`
	Action    string   `json:"action"`
}

func analyzeCallCost(start time.Time, end time.Time, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReaderSize(file, 4*units.MiB)
	orderedCall := make([]string, 0, 10)
	callCostMap := make(map[string][]costLog, 10)
	jobSyncInfo := make(map[int64]map[string]costLog, 10)
	collectEnd := end
	for {
		line, prefix, err2 := reader.ReadLine()
		if err2 != nil {
			if err2 == io.EOF {
				break
			}
			return err2
		}
		if prefix {
			return fmt.Errorf("line too long")
		}
		header := logHeader{}
		err2 = json.Unmarshal(line, &header)
		if err2 != nil {
			fmt.Println("skip invalid line: ", string(line))
			continue
		}
		if header.Time.Before(start) {
			continue
		}
		if !end.IsZero() && header.Time.After(end) {
			break
		}

		if header.Message == "DDL cost analysis" {
			collectEnd = header.Time.Time

			cl := costLog{}
			err2 = json.Unmarshal(line, &cl)
			if err2 != nil {
				return err2
			}
			callKey := cl.Call
			if cl.JobState != "" {
				callKey += "-" + cl.JobState
			}
			_, ok := callCostMap[callKey]
			if !ok {
				orderedCall = append(orderedCall, callKey)
				callCostMap[callKey] = make([]costLog, 0, 2000)
			}
			callCostMap[callKey] = append(callCostMap[callKey], cl)
		} else if header.Message == "SYNC VERSION for job" {
			cl := costLog{}
			err2 = json.Unmarshal(line, &cl)
			if err2 != nil {
				return err2
			}
			m, ok := jobSyncInfo[cl.JobID]
			if !ok {
				m = make(map[string]costLog, 2)
				jobSyncInfo[cl.JobID] = m
			}
			m[cl.Action] = cl
		}
	}
	fmt.Printf("analyze from %s to %s, duration: %s\n", start.Format(logTimeFormat),
		collectEnd.Format(logTimeFormat), collectEnd.Sub(start).Round(time.Millisecond).String())
	for _, call := range orderedCall {
		costs := callCostMap[call]
		durations := make([]time.Duration, 0, len(costs))
		for _, cost := range costs {
			durations = append(durations, cost.Cost.Duration)
		}
		printPercentile(call, durations)
	}
	gapDurations := make([]time.Duration, 0, len(jobSyncInfo))
	fmt.Println("jobSyncInfo len: ", len(jobSyncInfo))
	for _, jsi := range jobSyncInfo {
		writeLog, okw := jsi["update-global"]
		evLog, okb := jsi["receive-event"]
		if okw && okb {
			gapDurations = append(gapDurations, evLog.Time.Time.Sub(writeLog.Time.Time))
		}
	}
	if len(gapDurations) > 0 {
		printPercentile("write-pd-receive-gap", gapDurations)
	}
	return nil
}
