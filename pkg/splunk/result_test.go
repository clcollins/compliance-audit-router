package splunk

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestSearchResult_string(t *testing.T) {
	type args struct {
		field string
	}
	tests := []struct {
		name       string
		a          SearchResult
		args       args
		wantString string
		wantLog    string
	}{
		{
			name: "String field should pass",
			a: SearchResult{
				"alertname": "testAlertname",
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "alertname",
			},
			wantString: "testAlertname",
			wantLog:    "",
		},
		{
			// Maybe this SHOULDN'T pass, though
			name: "Empty string field should pass",
			a: SearchResult{
				"alertname": "",
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "alertname",
			},
			wantString: "",
			wantLog:    "",
		},
		{
			name: "Missing field should fail",
			a: SearchResult{
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "alertname",
			},
			wantString: "",
			wantLog:    "No such field: alertname\n",
		},
		{
			// This should probably fail too
			name: "Valid timestamp field should pass",
			a: SearchResult{
				"alertname": "testAlertname",
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "timestamp",
			},
			wantString: "2021-01-01T00:00:00.GMT",
			wantLog:    "",
		},
		{
			// Maybe this should fail as well
			name: "Slice field should pass with string representation of slice",
			a: SearchResult{
				"alertname": "testAlertname",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "clusterid",
			},
			wantString: "[testClusterID1 testClusterID2]",
			wantLog:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture the logger output
			var buff bytes.Buffer
			log.SetOutput(&buff)
			log.SetFlags(0)

			defer log.SetOutput(io.Discard)
			defer log.SetFlags(3)

			if got := tt.a.string(tt.args.field); got != tt.wantString {
				t.Errorf("SearchResult.string() = %v, want %v", got, tt.wantString)
			}

			if buff.String() != tt.wantLog {
				t.Errorf("SearchResult.string() Log Message = %v, want %v", buff.String(), tt.wantLog)
			}
		})
	}
}

func TestSearchResult_slice(t *testing.T) {
	type args struct {
		field string
	}
	tests := []struct {
		name            string
		a               SearchResult
		args            args
		wantStringSlice []string
		wantLog         string
	}{
		{
			name: "Valid slice should pass",
			a: SearchResult{
				"alertname": "testAlertname",
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "clusterid",
			},
			wantStringSlice: []string{"testClusterID1", "testClusterID2"},
			wantLog:         "",
		},
		{
			name: "Empty Slice should pass",
			a: SearchResult{
				"alertname": "testAlertname",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "clusterid",
			},
			wantStringSlice: []string(nil),
			wantLog:         "",
		},
		{
			name: "Missing Slice should return an empty slice",
			a: SearchResult{
				"alertname": "testAlertname",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "clusterid",
			},
			wantStringSlice: []string{},
			wantLog:         "",
		},
		{
			name: "String should become slice",
			a: SearchResult{
				"alertname": "testAlertname",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": "testClusterID1",
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "clusterid",
			},
			wantStringSlice: []string{"testClusterID1"},
			wantLog:         "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture the logger output
			var buff bytes.Buffer
			log.SetOutput(&buff)
			log.SetFlags(0)

			defer log.SetOutput(io.Discard)
			defer log.SetFlags(3)

			if got := tt.a.slice(tt.args.field); !reflect.DeepEqual(got, tt.wantStringSlice) {
				fmt.Printf("%t", !reflect.DeepEqual(got, tt.wantStringSlice))
				fmt.Printf("%+v", got)
				fmt.Printf("%+v", tt.wantStringSlice)
				t.Errorf("SearchResult.slice() = %v, want %v", got, tt.wantStringSlice)
			}

			if buff.String() != tt.wantLog {
				t.Errorf("SearchResult.string() Log Message = %v, want %v", buff.String(), tt.wantLog)
			}
		})
	}
}

func TestSearchResult_time(t *testing.T) {
	var testTimestamp, _ = time.Parse("2006-01-02T15:04:05.GMT", "2021-01-01T00:00:00.GMT")

	type args struct {
		field string
	}
	tests := []struct {
		name     string
		a        SearchResult
		args     args
		wantTime time.Time
		wantLog  string
	}{
		{
			name: "Valid timestamp should pass",
			a: SearchResult{
				"alertname": "testAlertname",
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "timestamp",
			},
			wantTime: testTimestamp,
			wantLog:  "",
		},
		{
			name: "Empty timestamp should return beginning of time",
			a: SearchResult{
				"alertname": "testAlertname",
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "timestamp",
			},
			wantTime: time.Time{},
			wantLog:  "",
		},
		{
			name: "Missing timestamp should fail",
			a: SearchResult{
				"alertname": "testAlertname",
				"username":  "testUsername",
				"group":     "testGroup",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "timestamp",
			},
			wantTime: time.Time{},
			wantLog:  "No such field: timestamp\n",
		},
		{
			name: "Wrong input should fail",
			a: SearchResult{
				"alertname": "testAlertname",
				"username":  "testUsername",
				"group":     "testGroup",
				"timestamp": "2021-01-01T00:00:00.GMT",
				"clusterid": []string{"testClusterID1", "testClusterID2"},
				"reason":    []string{"testReason1", "testReason2"},
			},
			args: args{
				field: "alertname",
			},
			wantTime: time.Time{},
			wantLog:  "Error parsing timestamp: parsing time \"testAlertname\" as \"2006-01-02T15:04:05.GMT\": cannot parse \"testAlertname\" as \"2006\"\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture the logger output
			var buff bytes.Buffer
			log.SetOutput(&buff)
			log.SetFlags(0)

			defer log.SetOutput(io.Discard)
			defer log.SetFlags(3)

			if got := tt.a.time(tt.args.field); !reflect.DeepEqual(got, tt.wantTime) {
				t.Errorf("SearchResult.time() = %v, want %v", got, tt.wantTime)
			}
			if buff.String() != tt.wantLog {
				t.Errorf("SearchResult.string() Log Message = %v, want %v", buff.String(), tt.wantLog)
			}
		})
	}
}
