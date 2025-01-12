package views

import (
	"encoding/json"
	"errors"
	"excel_project/dialects"
	"excel_project/models"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm/clause"
)

func ProcessExcelFile(file multipart.File) error {
	var rows [][]string
	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		log.Println("Failed to read file file:", err)
		return err
	}
	sheetname := xlsx.GetSheetName(0)
	fmt.Println("sheetname = ", sheetname)
	if rows, err = xlsx.GetRows(sheetname); err != nil {
		log.Println("Failed to parse Excel file:", err)
		return err
	}

	// Validate column headers

	fmt.Println("rows[0] = ", rows[0][0])
	if len(rows) < 1 || rows[0][0] != "first_name" || rows[0][1] != "last_name" || rows[0][2] != "company_name" || rows[0][3] != "address" || rows[0][4] != "city" || rows[0][5] != "county" || rows[0][6] != "postal" || rows[0][7] != "phone" || rows[0][8] != "email" || rows[0][9] != "web" {
		log.Println("Invalid column headers")
		return errors.New("invalid column headers")

	}
	var records []models.Users
	for _, row := range rows {
		record := models.Users{
			ID:          uuid.New(),
			FirstName:   row[0],
			LastName:    row[1],
			CompanyName: row[2],
			Address:     row[3],
			City:        row[4],
			County:      row[5],
			Postal:      row[6],
			Phone:       row[7],
			Email:       row[8],
			Web:         row[9],
		}
		records = append(records, record)
		if len(records) == 20 {
			if err := StoredData(records); err != nil {
				log.Println("failed DB insert :", err.Error())
				return err
			}
			records = nil
		}

	}

	return nil
}

func StoredData(records []models.Users) error {
	if conn, err := dialects.GetConnection(); err != nil {
		log.Println("Failed to connect DB")
		return err
	} else {
		if tx := conn.Debug().Model(models.Users{}).Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "email"}},
				DoUpdates: clause.AssignmentColumns([]string{"first_name", "last_name", "company_name", "address", "city", "web", "county", "postal", "phone"}),
			},
		).Create(&records); tx.Error != nil {
			log.Println("failed DB insert :", tx.Error)
			return tx.Error

		}
	}
	if data, err := dialects.RedisClient.Get("records"); err != nil || data == "" {
		jsonData, _ := json.Marshal(&records)
		go dialects.RedisClient.SetE("records", string(jsonData), time.Duration(5*time.Minute))
	}
	return nil

}

func GetAllRecords(limit int, offset int) ([]models.Users, error) {
	var records []models.Users
	if conn, err := dialects.GetConnection(); err != nil {
		log.Println("Failed to connect DB")
		return nil, err
	} else {
		if tx := conn.Debug().Model(&models.Users{}).Limit(limit).Offset(offset).Find(&records); tx.Error != nil {
			log.Println("Failed to get all records :", tx.Error)
			return nil, tx.Error
		} else {
			return records, nil
		}
	}
}

func GetSingleRecords(User *models.Users) (*models.Users, error) {
	if conn, err := dialects.GetConnection(); err != nil {
		log.Println("Failed to connect DB")
		return nil, err
	} else {
		if tx := conn.Debug().Model(&models.Users{}).First(&User); tx.Error != nil {
			log.Println("Failed to get all records :", tx.Error)
			return nil, tx.Error
		} else {
			return User, nil
		}
	}
}

func Update(user *models.Users) error {
	if conn, err := dialects.GetConnection(); err != nil {
		log.Println("Failed to connect DB")
		return err
	} else {
		if tx := conn.Debug().Model(&models.Users{}).Where("id =?", user.ID).Updates(&user).First(&user); tx.Error != nil {
			log.Println("Failed to get all records :", tx.Error)
			return tx.Error
		} else {
			key := fmt.Sprintf("record:%s", user.ID)
			jsonData, _ := json.Marshal(&user)
			go dialects.RedisClient.SetE(key, string(jsonData), time.Duration(2*time.Minute))
			return nil
		}
	}
}
