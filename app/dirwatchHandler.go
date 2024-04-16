package app

import (
	"dirwatcher/app/core"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type DirWatchHandler struct {
	DirWatchService core.DirWatcherService
}

func (d *DirWatchHandler) GetTasks(ctx *fiber.Ctx) error {

	resp, err := d.DirWatchService.GetAllTasks()
	if err != nil {
		return core.InternalServerError{
			Message: "Error while trying to get all tasks",
			Cause:   err,
		}
	}
	fmt.Print("Get All Tasks Response", resp)

	return ctx.JSON(resp)

}

func (d *DirWatchHandler) GetTaskById(ctx *fiber.Ctx) error {

	id := ctx.Params("taskId")

	resp, err := d.DirWatchService.GetTaskById(id)
	if err != nil {
		return err
	}
	fmt.Print("Get Task By Id Response", resp)

	return ctx.JSON(resp)

}
