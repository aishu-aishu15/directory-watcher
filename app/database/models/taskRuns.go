package models

type FileInfo struct {
	FilesAdded   string
	Files        string
	FilesDeleted string
	Size         int
}

type DirInfo struct {
	Path     string
	FileInfo *FileInfo
}
type Status string

const (
	IN_PROGRESS Status = "InProgress"
	SUCCESS     Status = "Success"
	FAILED      Status = "Failed"
)

type TaskRuns struct {
	Id                    string `gorm:"primaryKey"`
	SystemId              string //MAC address of the host system
	StartTime             string
	CreatedTime           string
	DirInfo               string
	TotalMagicStringCount int
	UpdatedTime           string
	EndTime               string
	Status                Status
	MagicString           string
}
