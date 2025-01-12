package controllers

import (
	"encoding/json"
	"excel_project/dialects"
	"excel_project/models"
	"excel_project/views"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// APIS
//1

func Uploadfile(c *gin.Context) {
	f, err := c.FormFile("file")
	if err != nil {
		fmt.Println("error = ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return

	}

	file, err := f.Open()

	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Failed to open or process the Excel file. Please ensure the file is valid.",
		})
		return

	}
	defer file.Close()

	extension := strings.ToLower(filepath.Ext(f.Filename))
	if extension != ".xlsx" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": fmt.Sprintf("Invalid file type '%s'. Only .xlsx files are allowed.", extension),
		})
		return

	}
	go views.ProcessExcelFile(file)
	c.JSON(http.StatusOK, gin.H{"message": "File is being processed"})

}

func GetRecords(c *gin.Context) {
	limit, err := strconv.Atoi(c.Query("limit"))
	fmt.Println("limit", limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
		return
	}

	if data, err := views.GetAllRecords(limit, offset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch record"})
		return
	} else {
		c.JSON(http.StatusOK, data)
	}

}

func GetSingleRecord(c *gin.Context) {
	var user models.Users
	id := c.Param("user_id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record id"})
	}
	var err error
	if user.ID, err = uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record id"})
		return

	}

	key := fmt.Sprintf("record:%s", user.ID)

	if data, err := dialects.RedisClient.Get(key); err == nil && data != "" {
		if err := json.Unmarshal([]byte(data), &user); err == nil {
			c.JSON(http.StatusOK, user)
			return
		}
	} else {
		if data, err := views.GetSingleRecords(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
			return
		} else {
			jsonData, _ := json.Marshal(&user)
			go dialects.RedisClient.SetE(key, string(jsonData), time.Duration(2*time.Minute))
			c.JSON(http.StatusOK, data)
		}
	}

}

func UpdateRecord(c *gin.Context) {
	var user models.Users

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if err := views.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record in MySQL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Record updated successfully"})

}
