package model

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/swkisdust/torrentremover/internal/utils"
)

func TestFilterTorrents(t *testing.T) {
	testTorrents := []*Torrent{
		{Name: "Movie A", Category: "Movies", Tags: []string{"hd", "2023"}, Status: StatusUploading, Trackers: toTorrentTrackers([]string{"tracker.a.com", "tracker.b.com"})},
		{Name: "TV Show B", Category: "TV Shows", Tags: []string{"4k", "series"}, Status: StatusDownloading, Trackers: toTorrentTrackers([]string{"tracker.c.com"})},
		{Name: "Game C", Category: "Games", Tags: []string{"pc", "rpg"}, Status: StatusPaused, Trackers: toTorrentTrackers([]string{"tracker.d.com"})},
		{Name: "Music D", Category: "Music", Tags: []string{"flac"}, Status: StatusUploading, Trackers: toTorrentTrackers([]string{"tracker.e.com"})},
		{Name: "Movie E", Category: "Movies", Tags: []string{"sd", "old"}, Status: StatusStopped | StatusUploading, Trackers: toTorrentTrackers([]string{"tracker.a.com", "tracker.f.com"})},
		{Name: "TV Show F", Category: "TV Shows", Tags: []string{"anime"}, Status: StatusUploading, Trackers: toTorrentTrackers([]string{"tracker.g.com"})},
		{Name: "TV Show G", Category: "TV Shows", Tags: []string{"hd"}, Status: StatusError, Trackers: toTorrentTrackers([]string{"tracker.a.com"})},
	}

	tests := []struct {
		name     string
		filter   Filters
		input    []*Torrent
		expected []*Torrent
	}{
		{
			name:     "No filter (should return all)",
			filter:   Filters{},
			input:    testTorrents,
			expected: testTorrents,
		},
		{
			name:     "Exclude Movies category",
			filter:   Filters{ExcludedCategories: []string{"Movies"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[5], testTorrents[6]},
		},
		{
			name:     "Exclude 'hd' tag",
			filter:   Filters{ExcludedTags: []string{"hd"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Exclude 'paused' status (using bit flag)",
			filter:   Filters{ExcludedStatus: []Status{StatusPaused}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[0], testTorrents[1], testTorrents[3], testTorrents[4], testTorrents[5], testTorrents[6]},
		},
		{
			name:     "Exclude 'error' status (using bit flag, affects combined status)",
			filter:   Filters{ExcludedStatus: []Status{StatusError}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[0], testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Exclude 'paused' AND 'error' status (any of these flags)",
			filter:   Filters{ExcludedStatus: []Status{StatusPaused, StatusError}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[0], testTorrents[1], testTorrents[3], testTorrents[4], testTorrents[5]},
		},
		{
			name:     "Exclude 'tracker.b.com' substring in trackers",
			filter:   Filters{ExcludedTrackers: []string{"tracker.b.com"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[1], testTorrents[2], testTorrents[3], testTorrents[4], testTorrents[5], testTorrents[6]},
		},
		{
			name:     "Include only 'TV Shows' category",
			filter:   Filters{Categories: []string{"TV Shows"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[1], testTorrents[5], testTorrents[6]},
		},
		{
			name:     "Include only '4k' tag",
			filter:   Filters{Tags: []string{"4k"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[1]},
		},
		{
			name:     "Include only 'uploading' status",
			filter:   Filters{Status: []Status{StatusUploading}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[0], testTorrents[3], testTorrents[4], testTorrents[5]},
		},

		{
			name:     "Include only 'downloading' status (using bit flag, affects combined status)",
			filter:   Filters{Status: []Status{StatusDownloading}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[1]},
		},
		{
			name:     "Include only 'paused' OR 'error' status",
			filter:   Filters{Status: []Status{StatusPaused, StatusError}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[2], testTorrents[6]},
		},
		{
			name:     "Include only 'tracker.c.com' substring in trackers",
			filter:   Filters{Trackers: []string{"tracker.c.com"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[1]},
		},
		{
			name:     "Include only 'tracker.a.com' substring AND 'anime' OR 'hd' tag in trackers",
			filter:   Filters{Trackers: []string{"tracker.a.com"}, Tags: []string{"anime", "hd"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[0], testTorrents[6]},
		},
		{
			name:     "Combined filter: Exclude 'Movies' AND Include only 'uploading' status",
			filter:   Filters{ExcludedCategories: []string{"Movies"}, Status: []Status{StatusUploading}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[3], testTorrents[5]},
		},
		{
			name:     "Combined filter: Exclude 'old' tag AND Include 'Movies' category",
			filter:   Filters{ExcludedTags: []string{"old"}, Categories: []string{"Movies"}},
			input:    testTorrents,
			expected: []*Torrent{testTorrents[0]},
		},
		{
			name:     "Disk filter: Has more remaining space",
			filter:   Filters{Disk: 1024},
			input:    testTorrents,
			expected: nil,
		},
		{
			name:     "Disk filter: Has less remaining space",
			filter:   Filters{Disk: 4096},
			input:    testTorrents,
			expected: testTorrents,
		},
		{
			name:     "Empty input torrents",
			filter:   Filters{Categories: []string{"Movies"}},
			input:    []*Torrent{},
			expected: []*Torrent{},
		},
		{
			name:     "No matching torrents for filter",
			filter:   Filters{Tags: []string{"nonexistent_tag"}},
			input:    testTorrents,
			expected: []*Torrent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterTorrents(&tt.filter, 2048, tt.input)

			sortTorrents(got)
			sortTorrents(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("FilterTorrents(%+v, %v) = \nGOT:  %v\nWANT: %v", tt.filter, tt.input, got, tt.expected)
			}
		})
	}
}

func sortTorrents(torrents []*Torrent) {
	slices.SortFunc(torrents, func(a, b *Torrent) int {
		return strings.Compare(a.Name, b.Name)
	})
}

func toTorrentTrackers(hosts []string) []TorrentTracker {
	return utils.SlicesMap(hosts, func(host string) TorrentTracker {
		return TorrentTracker{
			URL: host,
		}
	})
}
