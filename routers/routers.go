package routers

import (
	"excel_project/controllers"

	"github.com/gin-gonic/gin"
)

func Endpoints(routers *gin.Engine) {

	routers.POST("/upload", controllers.Uploadfile)
	routers.GET("/users", controllers.GetRecords)
	routers.GET("/user/:user_id", controllers.GetSingleRecord)
	routers.PUT("/user/:user_id", controllers.UpdateRecord)
	// routers.DELETE("/records/:record_id", controller.DeleteRecord)

}
