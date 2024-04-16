package app

import (
	"dirwatcher/app/core"

	"github.com/gofiber/fiber/v2"
)

type ConfigurationHandler struct {
	ConfigurationService core.ConfigurationService
}

func (c *ConfigurationHandler) CreateORUpdateConfig(ctx *fiber.Ctx) error {

	request := new(core.ConfigRequest)
	bodyParseErr := ctx.BodyParser(request)

	if bodyParseErr != nil {
		return core.InvalidRequestError{
			Message: "Bad Request",
			Cause:   bodyParseErr,
		}

	}

	resp, err := c.ConfigurationService.CreateORUpdateConfig(*request)
	if err != nil {
		return err
	}

	return ctx.JSON(resp)

}
