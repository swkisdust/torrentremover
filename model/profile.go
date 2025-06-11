package model

type Profile struct {
	Client   string     `json:"client"`
	Strategy []Strategy `json:"strategy"`
}
