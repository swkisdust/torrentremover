package model

import (
	"fmt"
	"strings"

	"slices"

	"github.com/swkisdust/torrentremover/internal/format"
)

type Strategy struct {
	Name        string `json:"name"`
	Filter      Filter `json:"filter"`
	Reannounce  bool   `json:"reannounce,omitempty"`
	DeleteFiles bool   `json:"delete_files,omitempty"`
	RemoveExpr  string `json:"remove"`
}

type Filter struct {
	Categories format.Array[string] `json:"categories,omitempty"`
	Tags       format.Array[string] `json:"tags,omitempty"`
	Trackers   format.Array[string] `json:"trackers,omitempty"`
	Status     format.Array[Status] `json:"status,omitempty"`

	ExcludedCategories format.Array[string] `json:"excluded_categories,omitempty"`
	ExcludedTags       format.Array[string] `json:"excluded_tags,omitempty"`
	ExcludedTrackers   format.Array[string] `json:"excluded_trackers,omitempty"`
	ExcludedStatus     format.Array[Status] `json:"excluded_status,omitempty"`
}

// Torrent status
type Status uint16

const (
	StatusDownloading Status = 1 << iota
	StatusUploading
	StatusError

	StatusPaused  Status = 1 << 12
	StatusQueued  Status = 1 << 13
	StatusStalled Status = 1 << 14
	StatusStopped Status = 1 << 15
)

func (s Status) HasFlag(flag Status) bool {
	return s&flag == flag
}

func (s *Status) UnmarshalYAML(b []byte) error {
	statusStr := string(b)
	code := GetStatus(statusStr)
	if code == 0 && statusStr != "" {
		return fmt.Errorf("invalid status string in YAML: %q", statusStr)
	}

	*s = code
	return nil
}

func GetStatus(s string) Status {
	switch strings.ToLower(s) {
	case "uploading":
		return StatusUploading
	case "downloading":
		return StatusDownloading
	case "paused":
		return StatusPaused
	case "queued":
		return StatusQueued
	case "stopped":
		return StatusStopped
	case "error":
		return StatusError
	case "pausedup":
		return StatusPaused | StatusUploading
	case "pauseddl":
		return StatusPaused | StatusDownloading
	case "queuedup":
		return StatusQueued | StatusUploading
	case "queueddl":
		return StatusQueued | StatusDownloading
	case "stoppedup":
		return StatusStopped | StatusUploading
	case "stoppeddl":
		return StatusStopped | StatusDownloading
	case "stalledup":
		return StatusStalled | StatusUploading
	case "stalleddl":
		return StatusStalled | StatusDownloading
	default:
		return 0
	}
}

func ContainStatus(s []Status, v Status) bool {
	if len(s) == 0 {
		return false
	}

	return slices.ContainsFunc(s, v.HasFlag)
}
