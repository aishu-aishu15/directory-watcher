package core

type ConfigRequest struct {
	Interval      string `json:"interval"`
	DirectoryPath string `json:"directoryPath"`
	MagicString   string `json:"magicString"`
}

type ConfigResponse struct {
	Id            string `json:"id"`
	Interval      string `json:"interval"`
	DirectoryPath string `json:"directoryPath"`
	MagicString   string `json:"magicString"`
}

type FileInfo struct {
	FilesAdded   string `json:"filesAdded"`
	Files        string `json:"files"`
	FilesDeleted string `json:"filesDeleted"`
}

type DirInfo struct {
	Path     string    `json:"path"`
	FileInfo *FileInfo `json:"fileInfo"`
}
type Status string

const (
	IN_PROGRESS Status = "InProgress"
	SUCCESS     Status = "Success"
	FAILED      Status = "Failed"
)

type TaskResponse struct {
	Id                    string   `json:"id"`
	SystemId              string   `json:"systemId"`
	StartTime             string   `json:"startTime"`
	CreatedTime           string   `json:"createdTime"`
	DirInfo               *DirInfo `json:"dirInfo"`
	TotalMagicStringCount int      `json:"totalMagicStringCount"`
	UpdatedTime           string   `json:"updatedTime"`
	EndTime               string   `json:"endTime"`
	Status                Status   `json:"status"`
	MagicString           string   `json:"magicString"`
}
