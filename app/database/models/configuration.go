package models

type Configuration struct {
	Id            string `gorm:"primaryKey"`
	Interval      string
	DirectoryPath string
	MagicString   string
}
