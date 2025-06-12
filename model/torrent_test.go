package model

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestFilterTorrents(t *testing.T) {
	testTorrents := []Torrent{
		{Name: "Movie A", Category: "Movies", Tags: []string{"hd", "2023"}, Status: StatusUploading, Trackers: []string{"tracker.a.com", "tracker.b.com"}},
		{Name: "TV Show B", Category: "TV Shows", Tags: []string{"4k", "series"}, Status: StatusDownloading, Trackers: []string{"tracker.c.com"}},
		{Name: "Game C", Category: "Games", Tags: []string{"pc", "rpg"}, Status: StatusPaused, Trackers: []string{"tracker.d.com"}},
		{Name: "Music D", Category: "Music", Tags: []string{"flac"}, Status: StatusUploading, Trackers: []string{"tracker.e.com"}},
		{Name: "Movie E", Category: "Movies", Tags: []string{"sd", "old"}, Status: StatusStopped | StatusUploading, Trackers: []string{"tracker.a.com", "tracker.f.com"}},
		{Name: "TV Show F", Category: "TV Shows", Tags: []string{"anime"}, Status: StatusUploading, Trackers: []string{"tracker.g.com"}},
	}

	tests := []struct {
		name     string
		filter   Filter
		input    []Torrent
		expected []Torrent
	}{
		{
			name:     "No filter (should return all)",
			filter:   Filter{},
			input:    testTorrents,
			expected: testTorrents,
		},
		{
			name:     "Exclude Movies category",
			filter:   Filter{ExcludedCategories: []string{"Movies"}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[5]},
		},
		{
			name:     "Exclude 'hd' tag",
			filter:   Filter{ExcludedTags: []string{"hd"}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Exclude 'paused' status (using bit flag)",
			filter:   Filter{ExcludedStatus: []Status{StatusPaused}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[0], testTorrents[1], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Exclude 'error' status (using bit flag, affects combined status)",
			filter:   Filter{ExcludedStatus: []Status{StatusError}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[0], testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Exclude 'paused' AND 'error' status (any of these flags)",
			filter:   Filter{ExcludedStatus: []Status{StatusPaused, StatusError}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[0], testTorrents[1], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Exclude 'tracker.b.com' substring in trackers",
			filter:   Filter{ExcludedTrackers: []string{"tracker.b.com"}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Include only 'TV Shows' category",
			filter:   Filter{Categories: []string{"TV Shows"}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[1], testTorrents[5]},
		},
		{
			name:     "Include only '4k' tag",
			filter:   Filter{Tags: []string{"4k"}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[1]},
		},
		{
			name:     "Include only 'uploading' status",
			filter:   Filter{Status: []Status{StatusUploading}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[0], testTorrents[3], testTorrents[4], testTorrents[5]},
		},

		{
			name:     "Include only 'downloading' status (using bit flag, affects combined status)",
			filter:   Filter{Status: []Status{StatusDownloading}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[1]},
		},
		{
			name:     "Include only 'paused' OR 'error' status",
			filter:   Filter{Status: []Status{StatusPaused, StatusError}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[2]},
		},
		{
			name:     "Include only 'tracker.c.com' substring in trackers",
			filter:   Filter{Trackers: []string{"tracker.c.com"}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[1]},
		},
		{
			name:     "Combined filter: Exclude 'Movies' AND Include only 'uploading' status",
			filter:   Filter{ExcludedCategories: []string{"Movies"}, Status: []Status{StatusUploading}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[3], testTorrents[5]},
		},
		{
			name:     "Combined filter: Exclude 'old' tag AND Include 'Movies' category",
			filter:   Filter{ExcludedTags: []string{"old"}, Categories: []string{"Movies"}},
			input:    testTorrents,
			expected: []Torrent{testTorrents[0]},
		},
		{
			name:     "Empty input torrents",
			filter:   Filter{Categories: []string{"Movies"}},
			input:    []Torrent{},
			expected: []Torrent{},
		},
		{
			name:     "No matching torrents for filter",
			filter:   Filter{Tags: []string{"nonexistent_tag"}},
			input:    testTorrents,
			expected: []Torrent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterTorrents(tt.filter, tt.input)

			sortTorrents(got)
			sortTorrents(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("FilterTorrents(%+v, %v) = \nGOT:  %v\nWANT: %v", tt.filter, tt.input, got, tt.expected)
			}
		})
	}
}

func sortTorrents(torrents []Torrent) {
	slices.SortFunc(torrents, func(a, b Torrent) int {
		return strings.Compare(a.Name, b.Name)
	})
}
