package splunk

import (
	"reflect"
	"testing"
	"time"
)

func TestNewAlertDetails(t *testing.T) {
	var testTimestamp, _ = time.Parse("2006-01-02T15:04:05.GMT", "2021-01-01T00:00:00.GMT")

	type args struct {
		result SearchResult
	}
	tests := []struct {
		name string
		args args
		want AlertDetails
	}{
		{
			name: "Valid results should pass",
			args: args{
				result: SearchResult{
					"alertname":             "testAlertname",
					"username":              "testUsername",
					"group":                 "testGroup",
					"timestamp":             "2021-01-01T00:00:00.GMT",
					"clusterid":             []string{"testClusterID1", "testClusterID2"},
					"cluster_text":          "testClusterID1, testClusterID2",
					"elevated_summary":      []string{"DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200"},
					"elevated_summary_text": "*COMMANDS*:\n - DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200",
					"reason":                []string{"testReason1", "testReason2"},
					"reason_text":           "*REASONS*:\n - testReason1\n - testReason2",
				},
			},
			want: AlertDetails{
				AlertName:           "testAlertname",
				User:                "testUsername",
				Group:               "testGroup",
				Timestamp:           testTimestamp,
				ClusterIDs:          []string{"testClusterID1", "testClusterID2"},
				ClusterText:         "testClusterID1, testClusterID2",
				ElevatedSummary:     []string{"DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200"},
				ElevatedSummaryText: "*COMMANDS*:\n - DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200",
				Reasons:             []string{"testReason1", "testReason2"},
				ReasonsText:         "*REASONS*:\n - testReason1\n - testReason2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAlertDetails(tt.args.result); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s\n\tgot:\n%+v\n\twant:\n%+v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAlert_Details(t *testing.T) {
	var testTimestamp, _ = time.Parse("2006-01-02T15:04:05.GMT", "2021-01-01T00:00:00.GMT")

	tests := []struct {
		name string
		w    Alert
		want []AlertDetails
	}{
		{
			name: "Valid slice of results should pass",
			w: Alert{
				SearchResults: SearchResults{
					Results: []SearchResult{
						{
							"alertname":             "testAlertname",
							"username":              "testUsername",
							"group":                 "testGroup",
							"timestamp":             "2021-01-01T00:00:00.GMT",
							"clusterid":             []string{"testClusterID1", "testClusterID2"},
							"cluster_text":          "testClusterID1, testClusterID2",
							"elevated_summary":      []string{"DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200"},
							"elevated_summary_text": "*COMMANDS*:\n - DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200",
							"reason":                []string{"testReason1", "testReason2"},
							"reason_text":           "*REASONS*:\n - testReason1\n - testReason2",
						},
						{
							"alertname":             "testAlertname",
							"username":              "testUsername",
							"group":                 "testGroup",
							"timestamp":             "2021-01-01T00:00:00.GMT",
							"clusterid":             []string{"testClusterID3", "testClusterID4"},
							"cluster_text":          "testClusterID3, testClusterID4",
							"elevated_summary":      []string{"PATCH secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200"},
							"elevated_summary_text": "*COMMANDS*:\n - PATCH /cluster-version-operator-serving-cert (openshift-cluster-version) 200",
							"reason":                []string{"testReason3", "testReason4"},
							"reason_text":           "*REASONS*:\n - testReason3\n - testReason4",
						},
					},
				},
			},
			want: []AlertDetails{
				{
					AlertName:           "testAlertname",
					User:                "testUsername",
					Group:               "testGroup",
					Timestamp:           testTimestamp,
					ClusterIDs:          []string{"testClusterID1", "testClusterID2"},
					ClusterText:         "testClusterID1, testClusterID2",
					ElevatedSummary:     []string{"DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200"},
					ElevatedSummaryText: "*COMMANDS*:\n - DELETE secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200",
					Reasons:             []string{"testReason1", "testReason2"},
					ReasonsText:         "*REASONS*:\n - testReason1\n - testReason2",
				},
				{
					AlertName:           "testAlertname",
					User:                "testUsername",
					Group:               "testGroup",
					Timestamp:           testTimestamp,
					ClusterIDs:          []string{"testClusterID3", "testClusterID4"},
					ClusterText:         "testClusterID3, testClusterID4",
					ElevatedSummary:     []string{"PATCH secrets/cluster-version-operator-serving-cert (openshift-cluster-version) 200"},
					ElevatedSummaryText: "*COMMANDS*:\n - PATCH /cluster-version-operator-serving-cert (openshift-cluster-version) 200",
					Reasons:             []string{"testReason3", "testReason4"},
					ReasonsText:         "*REASONS*:\n - testReason3\n - testReason4",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.w.Details(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s\n\tgot:\n%+v\n\twant:\n%+v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAlertDetails_Valid(t *testing.T) {
	var testTimestamp, _ = time.Parse("2006-01-02T15:04:05.GMT", "2021-01-01T00:00:00.GMT")

	tests := []struct {
		name string
		w    AlertDetails
		want bool
	}{
		{
			name: "Valid AlertDetails should pass",
			w: AlertDetails{
				AlertName:  "testAlertname",
				User:       "testUsername",
				Group:      "testGroup",
				Timestamp:  testTimestamp,
				ClusterIDs: []string{"testClusterID1", "testClusterID2"},
				Reasons:    []string{"testReason1", "testReason2"},
			},
			want: true,
		},
		{
			name: "Missing AlertName should fail",
			w: AlertDetails{
				AlertName:  "",
				User:       "testUsername",
				Group:      "testGroup",
				Timestamp:  testTimestamp,
				ClusterIDs: []string{"testClusterID1", "testClusterID2"},
				Reasons:    []string{"testReason1", "testReason2"},
			},
			want: false,
		},
		{
			name: "Missing user should fail",
			w: AlertDetails{
				AlertName:  "testAlertname",
				User:       "",
				Group:      "testGroup",
				Timestamp:  testTimestamp,
				ClusterIDs: []string{"testClusterID1", "testClusterID2"},
				Reasons:    []string{"testReason1", "testReason2"},
			},
			want: false,
		},
		{
			name: "Missing group should fail",
			w: AlertDetails{
				AlertName:  "testAlertname",
				User:       "tesUsername",
				Group:      "",
				Timestamp:  testTimestamp,
				ClusterIDs: []string{"testClusterID1", "testClusterID2"},
				Reasons:    []string{"testReason1", "testReason2"},
			},
			want: false,
		},
		{
			name: "Empty slice should fail",
			w: AlertDetails{
				AlertName:  "testAlertname",
				User:       "testUsername",
				Group:      "testGroups",
				Timestamp:  testTimestamp,
				ClusterIDs: []string{},
				Reasons:    []string{"testReason1", "testReason2"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.w.Valid(); got != tt.want {
				t.Errorf("%s\n\tgot:\n%+v\n\twant:\n%+v", tt.name, got, tt.want)
			}
		})
	}
}
