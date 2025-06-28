package model

type SessionStats struct {
	TotalDlSpeed int64 `expr:"total_dl_speed"`
	TotalUpSpeed int64 `expr:"total_up_speed"`
}
