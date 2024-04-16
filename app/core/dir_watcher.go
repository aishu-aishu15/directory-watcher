package core

import (
	"dirwatcher/app/database/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DirWatcherService struct {
	db *gorm.DB
}

func (d *DirWatcherService) ProcessTask(task *models.TaskRuns, oldFiles string) (*models.TaskRuns, error) {
	var dirInfo models.DirInfo

	if err := json.Unmarshal([]byte(task.DirInfo), &dirInfo); err != nil {
		return nil, err
	}

	occurrences, filesInDirectory, err := d.GetMagicStringOccurrencesAndFilesInDirectory(dirInfo.Path, task.MagicString)
	if err != nil {
		return nil, InternalServerError{
			Message: "Error occurred while trying to get magic string occurrence and file info",
			Cause:   err,
		}
	}

	task.DirInfo = StructToJsonString(
		models.DirInfo{
			Path: dirInfo.Path,
			FileInfo: &models.FileInfo{
				FilesAdded:   "",
				FilesDeleted: "",
				Files:        toJsonString(filesInDirectory),
			},
		},
	)
	task.UpdatedTime = time.Now().Format(time.RFC3339)
	task.Status = models.SUCCESS
	task.TotalMagicStringCount = occurrences

	if len(oldFiles) > 0 {
		filesAdded, filesDeleted, err := d.GetFilesAddedAndDeleted(oldFiles, filesInDirectory)
		fmt.Println("Files added and deleted", filesAdded, filesDeleted)
		if err != nil {
			return nil, err
		}
		task.DirInfo = StructToJsonString(
			models.DirInfo{
				Path: dirInfo.Path,
				FileInfo: &models.FileInfo{
					FilesAdded:   toJsonString(filesAdded),
					FilesDeleted: toJsonString(filesDeleted),
					Files:        toJsonString(filesInDirectory),
				},
			},
		)
	}

	fmt.Println("Final task object to save", task)

	return task, nil

}

func (d *DirWatcherService) GetMagicStringOccurrencesAndFilesInDirectory(directoryPath string, magicString string) (int, []string, error) {

	var occurrences int
	var files []string
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		// Check for errors
		if err != nil {
			fmt.Printf("Encountered an error: %v\n", err)
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read the file contents
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			return nil
		}

		// Search for the string in the file content
		fileOccurrences := strings.Count(string(content), magicString)
		occurrences += fileOccurrences

		if err == nil && !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	// Check for errors while walking the directory
	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
	}
	fmt.Println("Occurrences", occurrences, files)
	return occurrences, files, nil
}

func (d *DirWatcherService) GetById(Id string) (*models.TaskRuns, error) {
	var task = models.TaskRuns{
		Id: Id,
	}

	result := d.db.First(&task, "id = ?", Id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &task, nil

}

func (d *DirWatcherService) Create(task *models.TaskRuns) error {
	result := d.db.Create(task)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *DirWatcherService) Update(task *models.TaskRuns) error {
	fmt.Println("Task update....", task)
	result := d.db.Model(&models.TaskRuns{}).Where("id = ?", task.Id).Updates(task)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *DirWatcherService) GetBySystemId(SystemId string) (*models.TaskRuns, error) {
	var latestTask = models.TaskRuns{}

	// Query the database to get the latest task run by system ID
	result := d.db.Where("system_id = ?", SystemId).Order("start_time desc").First(&latestTask)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Println("No records found for the given SystemId.")
			return nil, nil // Return nil for both task and error
		}
		return nil, result.Error // Return the error if it's not a "record not found" error
	}

	return &latestTask, nil // Return the latest task run and no error
}

func (d *DirWatcherService) GetFilesAddedAndDeleted(oldFiles string, newFiles []string) ([]string, []string, error) {

	var files []string
	if err := json.Unmarshal([]byte(oldFiles), &files); err != nil {
		return []string{}, []string{}, err

	}

	filesAdded := getFilesAdded(files, newFiles)
	filesDeleted := getFilesDeleted(files, newFiles)
	return filesAdded, filesDeleted, nil

}

func getFilesAdded(oldFiles []string, newFiles []string) []string {
	addedFiles := []string{}
	fileMap := make(map[string]bool)
	for _, file := range oldFiles {
		fileMap[file] = true
	}

	for _, file := range newFiles {
		if _, exist := fileMap[file]; !exist {
			addedFiles = append(addedFiles, file)
		}
	}
	fmt.Println("Added Files", addedFiles)

	return addedFiles
}

func getFilesDeleted(oldFiles []string, newFiles []string) []string {
	fileMap := make(map[string]bool)
	deletedFiles := []string{}

	for _, file := range newFiles {
		fileMap[file] = true
	}

	for _, file := range oldFiles {
		if _, exist := fileMap[file]; !exist {
			deletedFiles = append(deletedFiles, file)
		}
	}

	fmt.Println("Deleted files", deletedFiles)
	return deletedFiles

}

func toJsonString(data []string) string {

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonData)

}

func (d *DirWatcherService) GetAllTasks() ([]*models.TaskRuns, error) {
	var items []*models.TaskRuns
	result := d.db.Model(&models.TaskRuns{}).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}

	return items, nil

}

func (d *DirWatcherService) GetTaskById(id string) (*TaskResponse, error) {
	var item *TaskResponse
	if err := d.db.Where("id = ?", id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NotFoundError{
				Message: "Record not found",
				Cause:   err,
			}
		} else {
			return nil, InternalServerError{
				Message: fmt.Sprintf("Error while trying to get task by id %v", id),
			}
		}
	}

	return item, nil

}

func StructToJsonString(data models.DirInfo) string {

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonData)

}

func NewDirWatcherService(
	db *gorm.DB,
) DirWatcherService {
	return DirWatcherService{
		db: db,
	}
}
