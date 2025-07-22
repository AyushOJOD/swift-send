package models

type FileMeta struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	TotalChunks int    `json:"total_chunks"`
}
