package core

import (
	"dirwatcher/app/database"
	"dirwatcher/app/database/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ConfigurationService struct {
	db *gorm.DB
}

func (c *ConfigurationService) CreateORUpdateConfig(request ConfigRequest) (*ConfigResponse, error) {

	Id, err := GetMacAddress()
	if err != nil {
		return nil, InternalServerError{
			Message: "Failed to get system id or MAC address",
		}
	}
	config := models.Configuration{
		Id:            *Id,
		Interval:      request.Interval,
		DirectoryPath: request.DirectoryPath,
		MagicString:   request.MagicString,
	}

	configResponse, err := c.GetById(*Id)
	resp := ConfigResponse{
		Id:            *Id,
		Interval:      request.Interval,
		DirectoryPath: request.DirectoryPath,
		MagicString:   request.MagicString,
	}
	if err != nil {
		err = c.Create(&config)
		if err != nil {
			return nil, InternalServerError{
				Message: "Error while trying to create config record in db",
			}
		} else {
			return &resp, nil
		}

	}

	if (configResponse != &models.Configuration{}) {
		err := c.Update(&config)
		if err != nil {
			return nil, InternalServerError{
				Message: "Error while trying to update config db",
				Cause:   err,
			}
		} else {
			fmt.Printf("Updated config object for %v", Id)
			return &resp, nil

		}

	}

	return nil, nil
}

func (c *ConfigurationService) GetById(id string) (*models.Configuration, error) {
	var config = models.Configuration{
		Id: id,
	}

	result := c.db.First(&config, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, database.NoSuchRecordError{
				Message: "Record not Found",
			}
		} else {
			return nil, database.RepoError{
				Message: "Error while trying to get config",
			}
		}
	}

	return &config, nil
}

func (c *ConfigurationService) Create(config *models.Configuration) error {
	result := c.db.Create(config)
	if result.Error != nil {
		return result.Error
	}
	return nil

}

func (c *ConfigurationService) Update(config *models.Configuration) error {
	result := c.db.Model(&models.Configuration{}).Where("id = ?", config.Id).Updates(config)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func NewConfigurationService(
	db *gorm.DB,
) ConfigurationService {
	return ConfigurationService{
		db: db,
	}
}
