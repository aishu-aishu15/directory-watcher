# Directory Watcher Task

This application performs a directory watcher task to monitor changes in a specified directory.

# API Documentation

This document provides information about the endpoints available in the Application.

| Route           | Method | Request Body                     | Sample Response                   | Description                    |
| --------------- | ------ | -------------------------------- | --------------------------------- | ------------------------------ |
| `/v1/config`    | POST   | `{"interval": "5m","directoryPath": "/home/aiswarya/gotest","magicString": "jobs"}` | `{"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Alice"}]}` | Create/Update config           |
| `/v1/tasks`     | GET    | N/A                              | ```json<br>[<br>    {<br>        "id": "0e8e7bde-fb67-11ee-83b5-00be431e2025",<br>        "systemId": "00:be:43:1e:20:25",<br>        "startTime": "2024-04-16T02:00:53+05:30",<br>        "createdTime": "2024-04-16T02:00:53+05:30",<br>        "dirInfo": null,<br>        "totalMagicStringCount": 6,<br>        "updatedTime": "2024-04-16T02:00:53+05:30",<br>        "endTime": "2024-04-16T02:01:53+05:30",<br>        "status": "Success",<br>        "magicString": "jobs"<br>    }<br>]``` | Get All tasks                  |
| `/v1/tasks/:taskId` | GET    | N/A                              | `{"id": "0e8e7bde-fb67-11ee-83b5-00be431e2025", "systemId": "00:be:43:1e:20:25", "startTime": "2024-04-16T02:00:53+05:30", "createdTime": "2024-04-16T02:00:53+05:30", "dirInfo": null, "totalMagicStringCount": 6, "updatedTime": "2024-04-16T02:00:53+05:30", "endTime": "2024-04-16T02:01:53+05:30", "status": "Success", "magicString": "jobs"}` | Get task by id                 |
| `/v1/stop`      | POST   | N/A                              | N/A                               | Stops the process              |





## Setup

1. Ensure all dependencies are downloaded by running `go mod tidy`.

## Configuration

To configure the directory watcher task, provide the following inputs:

1. **Directory Path**: The path to the directory to watch. For example: `"/home/user/gotest"`.
2. **Interval**: The schedule interval for the background task. Use the following pattern:
    - `"1h"`: One hour.
    - `"1m"`: One minute.
    - `"1s"`: One second.
3. **Magic String**: The string to find occurrences in the watched directory. For example: `"jobs"`.

## Note
Created a default config to start the process initially, Please add configuration 

## Usage

To start the directory watcher task, initiate the application. To stop the task, use the `"/shutdown"` endpoint or use keys (ctrl+c, ctrl+z).


## DB schema
## configurations DB
| Field         | Type         | Null | Key | Default | Extra |
|---------------|--------------|------|-----|---------|-------|
| Id            | varchar(255) | NO   | PRI | NULL    |       |
| Interval      | varchar(255) | YES  |     | NULL    |       |
| DirectoryPath | varchar(255) | YES  |     | NULL    |       |
| MagicString   | varchar(255) | YES  |     | NULL    |       |

## task_runs DB



