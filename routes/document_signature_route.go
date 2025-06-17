package routes

import (
	"github.com/PhanPhuc2609/be-sign-file/constants"
	"github.com/PhanPhuc2609/be-sign-file/controller"
	"github.com/PhanPhuc2609/be-sign-file/middleware"
	"github.com/PhanPhuc2609/be-sign-file/service"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func DocumentRoutes(route *gin.Engine, injector *do.Injector) {
	docController := do.MustInvoke[controller.DocumentController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	routes := route.Group("/api/documents")
	{
		routes.POST("/upload", middleware.Authenticate(jwtService), docController.UploadDocument)
		routes.POST("/verify", middleware.Authenticate(jwtService), docController.UploadAndVerifyDocument)
		routes.GET(":id", docController.GetDocumentByID)
		routes.GET("/user", middleware.Authenticate(jwtService), docController.GetDocumentsByUserID)
		routes.DELETE(":id", docController.DeleteDocument)
	}
}

func SignatureRoutes(route *gin.Engine, injector *do.Injector) {
	sigController := do.MustInvoke[controller.SignatureController](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	routes := route.Group("/api/signatures")
	{
		routes.POST("", middleware.Authenticate(jwtService), sigController.CreateSignature)
		routes.GET(":id", sigController.GetSignatureByID)
		routes.GET("/document/:doc_id", sigController.GetSignaturesByDocumentID)
		routes.DELETE(":id", sigController.DeleteSignature)
	}
}
