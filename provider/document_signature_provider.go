package provider

import (
	"github.com/PhanPhuc2609/be-sign-file/controller"
	"github.com/PhanPhuc2609/be-sign-file/repository"
	"github.com/PhanPhuc2609/be-sign-file/service"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func ProvideDocumentDependencies(injector *do.Injector, db *gorm.DB) {
	docRepo := repository.NewDocumentRepository(db)
	docService := service.NewDocumentService(docRepo, db)
	do.Provide(
		injector, func(i *do.Injector) (controller.DocumentController, error) {
			return controller.NewDocumentController(docService), nil
		},
	)
}

func ProvideSignatureDependencies(injector *do.Injector, db *gorm.DB) {
	sigRepo := repository.NewSignatureRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	userRepo := repository.NewUserRepository(db)
	sigService := service.NewSignatureService(sigRepo, docRepo, userRepo, db)
	do.Provide(
		injector, func(i *do.Injector) (controller.SignatureController, error) {
			return controller.NewSignatureController(sigService), nil
		},
	)
}
