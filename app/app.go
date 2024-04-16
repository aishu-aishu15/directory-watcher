package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"dirwatcher/app/core"
	"dirwatcher/app/database"
	"dirwatcher/app/database/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Application struct {
	fiberApp                *fiber.App
	configDb                *gorm.DB
	taskRunDb               *gorm.DB
	configHandler           ConfigurationHandler
	configService           core.ConfigurationService
	wg                      sync.WaitGroup
	quit                    chan os.Signal
	directoryWatcherService core.DirWatcherService
	directoryWatcherHandler DirWatchHandler
}

func NewApplication() *Application {
	return &Application{
		quit: make(chan os.Signal, 1),
	}
}

func (a *Application) Start() {

	a.InitDatabaseConnection()

	a.configService = core.NewConfigurationService(
		a.configDb,
	)

	a.configHandler = ConfigurationHandler{
		a.configService,
	}

	a.directoryWatcherService = core.NewDirWatcherService(
		a.taskRunDb,
	)
	a.directoryWatcherHandler = DirWatchHandler{
		a.directoryWatcherService,
	}

	a.wg.Add(1)

	go a.InitHttp()
	go a.PeriodicTask()
	a.Shutdown()
}

func (a *Application) PeriodicTask() {
	defer a.wg.Done()

	systemId, err := core.GetMacAddress()
	if err != nil {
		fmt.Printf("Error getting MAC address: %v", err)
		return
	}
	fmt.Printf("ID: %s", *systemId)
	// default config is created to ensure task to start
	defaultConfig := "default-config"

	for {

		interval, directoryPath, magicString, err := a.GetConfig(systemId)
		if err != nil {
			fmt.Print("Fetching default configuration which is created to start the task initially Please Configure one")
			interval, directoryPath, magicString, err = a.GetConfig(&defaultConfig)
			if err != nil {
				panic(err)
			}
		}

		var latestTask = &models.TaskRuns{}

		latestTask, err = a.directoryWatcherService.GetBySystemId(*systemId)
		if err != nil {
			fmt.Printf("Error While trying to get system id %v", latestTask)
		}
		var oldFiles string
		var dirInfo models.DirInfo

		if latestTask != nil {
			if latestTask.DirInfo != "" {
				if err := json.Unmarshal([]byte(latestTask.DirInfo), &dirInfo); err != nil {
					log.Fatal("un marshalling dir info failed")
				}
				if dirInfo.FileInfo != nil {
					oldFiles = dirInfo.FileInfo.Files

				}
			}

		}

		id, err := uuid.NewUUID()
		if err != nil {
			fmt.Println(errors.New("failed to generate request id"), err)
		}

		newTask := models.TaskRuns{
			Id:          id.String(),
			SystemId:    *systemId,
			StartTime:   time.Now().Format(time.RFC3339),
			CreatedTime: time.Now().Format(time.RFC3339),
			DirInfo: StructToJsonString(models.DirInfo{
				Path: *directoryPath,
				FileInfo: &models.FileInfo{
					FilesAdded:   "",
					FilesDeleted: "",
					Files:        "",
				},
			}),
			Status:      models.IN_PROGRESS,
			MagicString: *magicString,
		}
		err = a.directoryWatcherService.Create(&newTask)
		if err != nil {
			fmt.Println("Error while trying to create task", newTask)
			panic(err)
		}

		task, err := a.directoryWatcherService.ProcessTask(&newTask, oldFiles)
		if err != nil {
			newTask.Status = models.FAILED
			newTask.UpdatedTime = time.Now().Format(time.RFC3339)

			err = a.directoryWatcherService.Update(&newTask)
			if err != nil {
				fmt.Printf("Error occurred while trying to update [%v]", err)
			}
		}

		err = a.directoryWatcherService.Update(task)
		if err != nil {
			fmt.Printf("Error occurred while trying to update [%v]", err)
		}

		// Task logic goes here
		fmt.Println("Interval to Perform", time.Duration(*interval)*time.Second)

		// Sleep for the specified interval before the next execution
		time.Sleep(time.Duration(*interval) * time.Second)

		task.EndTime = time.Now().Format(time.RFC3339)

		updateTaskErr := a.directoryWatcherService.Update(task)
		if updateTaskErr != nil {
			fmt.Printf("Error occurred while trying to update [%v]", err)

		}

	}
}

func (a *Application) GetConfig(Id *string) (*int, *string, *string, error) {
	config, err := a.configService.GetById(*Id)

	if err != nil {
		if errors.Is(err, database.NoSuchRecordError{}) {
			return nil, nil, nil, core.NotFoundError{
				Cause: err,
			}
		} else {
			return nil, nil, nil, core.InternalServerError{
				Cause: err,
			}
		}
	}

	log.Printf("Interval: %s", config.Interval)

	intervalInSeconds, err := GetIntervalInSeconds(config.Interval)
	if err != nil {
		return nil, nil, nil, core.InvalidRequestError{
			Message: "Enter valid interval refer README.md to know the allowed values",
		}
	}
	return &intervalInSeconds, &config.DirectoryPath, &config.MagicString, nil
}

func (a *Application) InitHttp() {

	a.fiberApp = fiber.New()
	a.registerAccountRoutes(a.fiberApp.Group("/v1"))

	err := a.fiberApp.Listen(":8000")
	if err != nil {
		panic(err)
	}

}

func (a *Application) registerAccountRoutes(router fiber.Router) {

	router.Post("/config", a.configHandler.CreateORUpdateConfig)
	router.Get("/tasks", a.directoryWatcherHandler.GetTasks)
	router.Get("/tasks/:taskId", a.directoryWatcherHandler.GetTaskById)
	router.Post("/stop", a.ShutDownHandler)

}

func (a *Application) InitDatabaseConnection() {

	configDB, err := gorm.Open(sqlite.Open("configurations.db"), &gorm.Config{})

	if err != nil {
		panic("failed to connect to database")
	}

	taskRunDB, err := gorm.Open(sqlite.Open("taskRuns.Db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	configDB.AutoMigrate(models.Configuration{})
	taskRunDB.AutoMigrate(models.TaskRuns{})

	a.configDb = configDB
	a.taskRunDb = taskRunDB
}

func GetIntervalInSeconds(interval string) (int, error) {

	durationSecondsMap := map[string]int{
		"h": 3600,
		"m": 60,
		"s": 1,
	}
	re := regexp.MustCompile(`^\d+`)
	match := re.FindString(interval)
	matchToInt := ConvertStrToInt(match)

	if strings.Contains(interval, "h") {
		return matchToInt * durationSecondsMap["h"], nil
	} else if strings.Contains(interval, "m") {
		return matchToInt * durationSecondsMap["m"], nil

	} else if strings.Contains(interval, "s") {
		return matchToInt * durationSecondsMap["s"], nil
	}

	return matchToInt * durationSecondsMap["h"], nil
}

func ConvertStrToInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return num
}

func (a *Application) Shutdown() {
	signal.Notify(a.quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-a.quit

	// Send quit signal to the periodic task goroutine
	close(a.quit)
	a.wg.Wait()
	var configdb *sql.DB
	var taskrundb *sql.DB
	var err error

	// Close database connections
	if configdb, err = a.configDb.DB(); err != nil {
		log.Printf("Error closing configuration database connection: %v", err)
	}
	configdb.Close()

	if taskrundb, err = a.taskRunDb.DB(); err != nil {
		log.Printf("Error closing task run database connection: %v", err)
	}
	taskrundb.Close()

	log.Println("Shutting down...")
	os.Exit(0)
}

func (a *Application) listenForShutdown(shutdown chan struct{}) {
	signal.Notify(a.quit, syscall.SIGINT, syscall.SIGTERM)
	<-a.quit

	close(shutdown)
	var configdb *sql.DB
	var taskrundb *sql.DB
	var err error

	// Close database connections
	if configdb, err = a.configDb.DB(); err != nil {
		log.Printf("Error closing configuration database connection: %v", err)
	}
	configdb.Close()

	if taskrundb, err = a.taskRunDb.DB(); err != nil {
		log.Printf("Error closing task run database connection: %v", err)
	}
	taskrundb.Close()

	log.Println("Shutting down...")
	os.Exit(0)
}

func (a *Application) ShutDownHandler(ctx *fiber.Ctx) error {
	shutdown := make(chan struct{})
	a.quit <- syscall.SIGINT
	go a.listenForShutdown(shutdown)
	a.wg.Wait()

	return nil
}

func StructToJsonString(data models.DirInfo) string {

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonData)

}
