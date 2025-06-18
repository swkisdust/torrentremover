package model

import (
	"fmt"
	"strings"

	"slices"

	"github.com/hekmon/transmissionrpc/v3"
	"github.com/swkisdust/torrentremover/internal/format"
	"github.com/swkisdust/torrentremover/internal/utils"
)

type Strategy struct {
	Name        string `json:"name"`
	Filter      Filter `json:"filter"`
	Reannounce  bool   `json:"reannounce,omitempty"`
	DeleteFiles bool   `json:"delete_files,omitempty"`
	DeleteDelay uint32 `json:"delete_delay,omitempty"`
	Mountpath   string `json:"mount_path,omitempty"`
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

	Disk Bytes `json:"disk,omitempty"`
}

// Torrent status
type Status uint16

const (
	StatusDownloading Status = 1 << iota
	StatusUploading
	StatusError
	StatusChecking

	StatusPaused  Status = 1 << 12
	StatusQueued  Status = 1 << 13
	StatusStalled Status = 1 << 14
	StatusStopped Status = 1 << 15
)

func (s Status) HasFlag(flag Status) bool {
	return s&flag == flag
}

func (s *Status) UnmarshalYAML(b []byte) error {
	statusStr := strings.Trim(string(b), `"`)
	code := GetStatus(statusStr)
	if code == 0 && statusStr != "" {
		return fmt.Errorf("invalid status string in YAML: %q", statusStr)
	}

	*s = code
	return nil
}

func GetStatus(s string) Status {
	switch strings.ToLower(s) {
	case "uploading", "seeding":
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
	case "checking":
		return StatusChecking
	// qbittorrent status
	case "pausedup", "pausedupload":
		return StatusPaused | StatusUploading
	case "pauseddl", "pauseddownload":
		return StatusPaused | StatusDownloading
	case "queuedup", "queuedupload":
		return StatusQueued | StatusUploading
	case "queueddl", "queueddownload":
		return StatusQueued | StatusDownloading
	case "stoppedup", "stoppedupload":
		return StatusStopped | StatusUploading
	case "stoppeddl", "stoppeddownload":
		return StatusStopped | StatusDownloading
	case "stalledup", "stalledupload":
		return StatusStalled | StatusUploading
	case "stalleddl", "stalleddownload":
		return StatusStalled | StatusDownloading
	case "checkingup", "checkingupload":
		return StatusChecking | StatusUploading
	case "checkingdl", "checkingdownload":
		return StatusChecking | StatusDownloading
	default:
		return 0
	}
}

func FromTrStatus(status transmissionrpc.TorrentStatus) Status {
	switch status {
	case transmissionrpc.TorrentStatusStopped:
		return StatusStopped
	case transmissionrpc.TorrentStatusDownloadWait:
		return StatusQueued | StatusDownloading
	case transmissionrpc.TorrentStatusDownload:
		return StatusDownloading
	case transmissionrpc.TorrentStatusSeedWait:
		return StatusQueued | StatusUploading
	case transmissionrpc.TorrentStatusSeed:
		return StatusUploading
	case transmissionrpc.TorrentStatusIsolated:
		return StatusStalled
	case transmissionrpc.TorrentStatusCheck:
		return StatusChecking
	case transmissionrpc.TorrentStatusCheckWait:
		return StatusQueued | StatusChecking
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

type Bytes int64

func (b *Bytes) UnmarshalYAML(buf []byte) error {
	bytesStr := strings.Trim(string(buf), `"`)
	bytes, err := utils.ParseBytes(bytesStr)
	if err != nil {
		return err
	}

	*b = Bytes(bytes)
	return nil
}
