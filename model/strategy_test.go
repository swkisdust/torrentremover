package model

import "testing"

func TestStatus(t *testing.T) {
	testSt := []struct {
		raw      string
		expected Status
		isUpload bool
	}{
		{
			raw:      "Downloading",
			expected: StatusDownloading,
			isUpload: false,
		},
		{
			raw:      "stalledUP",
			expected: StatusStalled | StatusUploading,
			isUpload: true,
		},
		{
			raw:      "paused",
			expected: StatusPaused,
			isUpload: false,
		},
		{
			raw:      "StoppedUpload",
			expected: StatusStopped | StatusUploading,
			isUpload: true,
		},
	}

	for _, st := range testSt {
		status := GetStatus(st.raw)
		if status != st.expected {
			t.Errorf("expected %v, got %v", st.expected, status)
		}
		if status.HasFlag(StatusUploading) != st.isUpload {
			t.Errorf("mask %v got wrong status", status)
		}
	}
}
