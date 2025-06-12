package model

type Profile struct {
	Client      string     `json:"client"`
	Strategy    []Strategy `json:"strategy"`
	Reannounce  bool       `json:"reannounce,omitempty"`
	DeleteFiles bool       `json:"delete_files,omitempty"`
}
