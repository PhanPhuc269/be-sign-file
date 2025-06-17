package migrations

import (
	"github.com/PhanPhuc2609/be-sign-file/entity"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
		&entity.Document{},
		&entity.Signature{},
	); err != nil {
		return err
	}

	return nil
}
