package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"hackitbackend/app/models"
	htmlTempalte "hackitbackend/app/template"
	"hackitbackend/pkg/utils"
	"hackitbackend/platform/database"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.Background()

func CreateUsers(c *fiber.Ctx) error {
	users := new(models.Users)

	if err := c.BodyParser(users); err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 400, err.Error(), nil)
	}

	// Generate a new UUID and check if it already exists
	for {
		users.ID = uuid.New().String()

		// Check if the UUID already exists
		existingUser := new(models.Users)
		err := database.Db.Collection("users").FindOne(ctx, bson.M{"_id": users.ID}).Decode(existingUser)

		if err == mongo.ErrNoDocuments {
			// UUID doesn't exist, we can use it
			break
		} else if err != nil {
			fmt.Println(err.Error())
			// An error occurred during the database query
			return utils.ResponseMsg(c, 500, "Error checking user ID", err.Error())
		}
		// If we reach here, the UUID already exists, so we'll generate a new one in the next iteration
	}

	users.CreatedAt = time.Now()
	users.UpdatedAt = time.Now()
	var err error
	users.VerificationCode = utils.GenerateVerificationCode()

	_, err = database.Db.Collection("users").InsertOne(ctx, users)
	if err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 500, "Failed to insert data", err.Error())
	}

	// TODO: Send verification email with the code

	return utils.ResponseMsg(c, 200, "User created successfully. Check email for verification code.", users)
}

func checkBase64Size(base64Str string, maxSize int) error {
	// 获取 Base64 字符串长度
	base64Length := len(base64Str)

	// 计算原始数据大小
	originalSize := (base64Length * 3) / 4

	// 减去填充字符的影响
	padding := strings.Count(base64Str, "=")
	originalSize -= padding

	// 检查是否超过最大大小
	if originalSize > maxSize {
		return errors.New("file size exceeds limit")
	}

	return nil
}

func CreateTeam(c *fiber.Ctx) error {
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	validate := validator.New()

	formData := new(struct {
		TeamName            string                      `json:"teamName" validate:"required,min=2,max=30"`
		TeamMembers         []models.Users              `json:"teamMembers" validate:"required,min=1,dive"`
		AccompanyingPersons []models.AccompanyingPerson `json:"accompanyingPersons" validate:"max=2,dive"`
		Exhibitors          []models.Exhibitor          `json:"exhibitors,omitempty" validate:"dive"`
	})

	// 解析請求的 JSON 數據
	if err := c.BodyParser(formData); err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// 驗證數據
	if err := validate.Struct(formData); err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	haveRepresentative := 0
	for _, member := range formData.TeamMembers {
		if member.IsRepresentative {
			haveRepresentative++
			break
		}
	}

	const maxFileSize = 10000000 // 10 MB

	for _, member := range formData.TeamMembers {
		if err := checkBase64Size(member.StudentCardFront, maxFileSize); err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Student card front is too large",
			})
		}

		// 检查学生证后面的 Base64 字符串大小
		if err := checkBase64Size(member.StudentCardBack, maxFileSize); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Student card back is too large",
			})
		}
	}

	// if len(formData.TeamMembers) < 3 || len(formData.TeamMembers) > 6 {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "The number of team members must be between 3 and 6",
	// 	})
	// }

	// 將 TeamMember、AccompanyingPerson 和 Exhibitor 實例插入資料庫並獲取其 UUID
	var teamMemberIDs []string
	teamID := uuid.New().String()
	for i, member := range formData.TeamMembers {
		username := member.Name
		if i == 0 {
			member.IsRepresentative = true
		} else {
			member.IsRepresentative = false
		}
		member.ID = uuid.New().String()
		member.VerificationCode = utils.GenerateVerificationCode()
		member.CheckedIn = false
		member.CreatedAt = time.Now()
		member.UpdatedAt = time.Now()

		// 生成128字元的隨機字符串
		randomStr, err := utils.GenerateRandomString(128)
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate random string",
			})
		}

		// 新的集合儲存 userID 和隨機字符串
		userVerification := &models.UserVerification{
			UserID: member.ID,
			Secret: randomStr,
		}
		_, err = database.Db.Collection("user_verifications_secret").InsertOne(c.Context(), userVerification)
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save random string for user",
			})
		}
		// 生成驗證 URL
		baseURL := os.Getenv("BASE_URL") + "/users/verification/"
		verificationURL := fmt.Sprintf("%s%s", baseURL, randomStr)

		for i, contact := range member.EmergencyContacts {
			encryptedPhone := utils.EncryptAES(contact.Phone, encryptionKey)
			member.EmergencyContacts[i].Phone = encryptedPhone
			encryptedRelationship := utils.EncryptAES(contact.Relationship, encryptionKey)
			member.EmergencyContacts[i].Relationship = encryptedRelationship
			encryptedName := utils.EncryptAES(contact.Name, encryptionKey)
			member.EmergencyContacts[i].Name = encryptedName
		}
		encryptedStudentCardFront := utils.EncryptAES(member.StudentCardFront, encryptionKey)
		member.StudentCardFront = encryptedStudentCardFront
			
		encryptedStudentCardBack := utils.EncryptAES(member.StudentCardBack, encryptionKey)
		member.StudentCardBack = encryptedStudentCardBack

		phone := utils.EncryptAES(member.Phone, encryptionKey)
		member.Phone = phone
		birthday := utils.EncryptAES(member.Birthday, encryptionKey)
		member.Birthday = birthday
		identityNumber := utils.EncryptAES(member.IdentityNumber, encryptionKey)
		member.IdentityNumber = identityNumber

		school := utils.EncryptAES(member.School, encryptionKey)
		member.School = school
		name := utils.EncryptAES(member.Name, encryptionKey)
		member.Name = name
		remarks := utils.EncryptAES(member.Remarks, encryptionKey)
		member.Remarks = remarks
		specialDiseases := utils.EncryptAES(member.SpecialDiseases, encryptionKey)
		member.SpecialDiseases = specialDiseases
		allergies := utils.EncryptAES(member.Allergies, encryptionKey)
		member.Allergies = allergies
		// 插入 team member
		_, err = database.Db.Collection("users").InsertOne(c.Context(), member)
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create team member",
			})
		}

		teamMemberIDs = append(teamMemberIDs, member.ID)

		// 如果成員是團隊代表人，則生成編輯表單的密鑰
		if member.IsRepresentative {
			editSecretString, err := utils.GenerateRandomString(128)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to generate random string",
				})
			}
			editSecret := &models.EditSecret{
				TeamID: teamID,
				Secret: editSecretString,
			}
			_, err = database.Db.Collection("edit_secrets").InsertOne(c.Context(), editSecret)
			if err != nil {
				fmt.Println(err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create edit secret",
				})
			}
			teamLeaderEditBaseURL := os.Getenv("BASE_URL_FRONTEND") + "/edit/"
			teamLeaderEditLink := fmt.Sprintf("%s%s", teamLeaderEditBaseURL, editSecretString)

			t, err := template.New("edit_form").Parse(htmlTempalte.TeamLeaderEditTemplate)
			if err != nil {
				fmt.Println(err.Error())
				return fmt.Errorf("failed to parse email template: %w", err)
			}

			var body bytes.Buffer
			type EmailData struct {
				Name     string
				EditLink string
			}

			if err := t.Execute(&body, EmailData{Name: username, EditLink: teamLeaderEditLink}); err != nil {
				fmt.Println(err.Error())
				return fmt.Errorf("failed to execute email template: %w", err)
			}

			// 發送編輯郵件
			err = utils.SendEmail(member.Email, "[Hackit] 對表單進行編輯", body.String())
			if err != nil {
				fmt.Println(err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to send email to %s: %v", member.Email, err),
				})
			}
		}

		t, err := template.New("email_verification").Parse(htmlTempalte.VerificationTemplate)
		if err != nil {
			fmt.Println(err.Error())
			return fmt.Errorf("failed to parse email template: %w", err)
		}

		var body bytes.Buffer
		type EmailData struct {
			Name             string
			VerificationLink string
		}

		if err := t.Execute(&body, EmailData{Name: username, VerificationLink: verificationURL}); err != nil {
			fmt.Println(err.Error())
			return fmt.Errorf("failed to execute email template: %w", err)
		}

		// 發送驗證郵件
		err = utils.SendEmail(member.Email, "[Hackit] 驗證您的郵件", body.String())
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to send email to %s: %v", member.Email, err),
			})
		}
	}

	var accompanyingPersonIDs []string
	for _, person := range formData.AccompanyingPersons {
		person.ID = uuid.New().String()
		encryptedPhone := utils.EncryptAES(person.Phone, encryptionKey)
		person.Phone = encryptedPhone
		_, err := database.Db.Collection("accompanying_persons").InsertOne(c.Context(), person)
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create accompanying person",
			})
		}
		accompanyingPersonIDs = append(accompanyingPersonIDs, person.ID)
	}

	var exhibitorIDs []string
	for _, exhibitor := range formData.Exhibitors {
		exhibitor.ID = uuid.New().String()
		_, err := database.Db.Collection("exhibitors").InsertOne(c.Context(), exhibitor)
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create exhibitor",
			})
		}
		exhibitorIDs = append(exhibitorIDs, exhibitor.ID)
	}

	// 準備將表單數據轉換為 Team 實例
	team := &models.Team{
		ID:                  teamID,
		TeamName:            formData.TeamName,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		TeamMembers:         teamMemberIDs,
		AccompanyingPersons: accompanyingPersonIDs,
		Exhibitors:          exhibitorIDs,
	}

	// 插入 Team 到 MongoDB
	collection := database.Db.Collection("teams")
	_, err := collection.InsertOne(c.Context(), team)
	if err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 500, "Failed to create team", err.Error())
	}

	return utils.ResponseMsg(c, 200, "Team created successfully", team)
}

func Verification(c *fiber.Ctx) error {
	// 取得 URL 中的 secret 參數
	secret := c.Params("secret")
	if secret == "" {
		return utils.ResponseMsg(c, 400, "Missing verification secret", nil)
	}

	// 在 user_verifications_secret 集合中尋找與 secret 相對應的 userID
	var userVerification models.UserVerification
	err := database.Db.Collection("user_verifications_secret").FindOne(c.Context(), bson.M{"secret": secret}).Decode(&userVerification)
	if err != nil {
		fmt.Println(err.Error())
		if err == mongo.ErrNoDocuments {
			return utils.ResponseMsg(c, 404, "Verification link is invalid or has expired", nil)
		}
		return utils.ResponseMsg(c, 500, "Error checking verification secret", err.Error())
	}

	// 根據找到的 userID 更新使用者的驗證狀態
	userNumber, err := utils.GetNextUserID(c.Context(), "user_number")
	if err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 500, "Failed to get next user number", err.Error())
	}
	filter := bson.M{"id": userVerification.UserID}
	update := bson.M{"$set": bson.M{"verified": true, "userNumber": userNumber}}
	_, err = database.Db.Collection("users").UpdateOne(c.Context(), filter, update)
	if err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 500, "Failed to update user verification status", err.Error())
	}

	_, err = database.Db.Collection("user_verifications_secret").DeleteOne(c.Context(), bson.M{"secret": secret})
	if err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 500, "Failed to delete verification token", err.Error())
	}

	return c.Redirect(os.Getenv("BASE_URL_FRONTEND")+"/verification/success", fiber.StatusSeeOther)
}

func GetFormInformation(c *fiber.Ctx) error {
	secret := c.Params("secret")
	encryptionKey := os.Getenv("ENCRYPTION_KEY")

	// 根據 secret 查詢 edit_secrets 集合
	var editSecret models.EditSecret
	err := database.Db.Collection("edit_secrets").FindOne(c.Context(), bson.M{"secret": secret}).Decode(&editSecret)
	if err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Invalid secret",
		})
	}

	// 根據 teamID 查詢團隊
	var team models.Team
	err = database.Db.Collection("teams").FindOne(c.Context(), bson.M{"id": editSecret.TeamID}).Decode(&team)
	if err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Team not found",
		})
	}

	// 查詢團隊成員
	var teamMembers []models.Users
	cursor, err := database.Db.Collection("users").Find(c.Context(), bson.M{"id": bson.M{"$in": team.TeamMembers}})
	if err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve team members",
		})
	}
	if err := cursor.All(c.Context(), &teamMembers); err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode team members",
		})
	}

	// 查詢 accompanying persons
	var accompanyingPersons []models.AccompanyingPerson
	if len(team.AccompanyingPersons) > 0 {
		cursor, err := database.Db.Collection("accompanying_persons").Find(c.Context(), bson.M{"id": bson.M{"$in": team.AccompanyingPersons}})
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve accompanying persons",
			})
		}
		if err := cursor.All(c.Context(), &accompanyingPersons); err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to decode accompanying persons",
			})
		}
	}

	// 查詢 exhibitors
	var exhibitors []models.Exhibitor
	if len(team.Exhibitors) > 0 {
		cursor, err := database.Db.Collection("exhibitors").Find(c.Context(), bson.M{"id": bson.M{"$in": team.Exhibitors}})
		if err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve exhibitors",
			})
		}
		if err := cursor.All(c.Context(), &exhibitors); err != nil {
			fmt.Println(err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to decode exhibitors",
			})
		}
	}


	for i, member := range teamMembers {
		for j, contact := range member.EmergencyContacts {
			encryptedPhone := utils.Decrypt(contact.Phone, encryptionKey)
			teamMembers[i].EmergencyContacts[j].Phone = encryptedPhone
			encryptedRelationship := utils.Decrypt(contact.Relationship, encryptionKey)
			teamMembers[i].EmergencyContacts[j].Relationship = encryptedRelationship
			encryptedName := utils.Decrypt(contact.Name, encryptionKey)
			teamMembers[i].EmergencyContacts[j].Name = encryptedName
		}
		encryptedStudentCardFront := utils.Decrypt(member.StudentCardFront, encryptionKey)
		teamMembers[i].StudentCardFront = encryptedStudentCardFront

		encryptedStudentCardBack := utils.Decrypt(member.StudentCardBack, encryptionKey)
		teamMembers[i].StudentCardBack = encryptedStudentCardBack

		phone := utils.Decrypt(member.Phone, encryptionKey)
		teamMembers[i].Phone = phone
		birthday := utils.Decrypt(member.Birthday, encryptionKey)
		teamMembers[i].Birthday = birthday
		identityNumber := utils.Decrypt(member.IdentityNumber, encryptionKey)
		teamMembers[i].IdentityNumber = identityNumber

		school := utils.Decrypt(member.School, encryptionKey)
		teamMembers[i].School = school
		name := utils.Decrypt(member.Name, encryptionKey)
		teamMembers[i].Name = name
		remarks := utils.Decrypt(member.Remarks, encryptionKey)
		teamMembers[i].Remarks = remarks
		specialDiseases := utils.Decrypt(member.SpecialDiseases, encryptionKey)
		teamMembers[i].SpecialDiseases = specialDiseases
		allergies := utils.Decrypt(member.Allergies, encryptionKey)
		teamMembers[i].Allergies = allergies
	}

	for i, person := range accompanyingPersons {
		phone := utils.Decrypt(person.Phone, encryptionKey)
		accompanyingPersons[i].Phone = phone
	}

	// 返回完整的團隊信息
	teamData := fiber.Map{
		"id":                  team.ID,
		"teamName":            team.TeamName,
		"teamMembers":         teamMembers,
		"accompanyingPersons": accompanyingPersons,
		"exhibitors":          exhibitors,
	}

	return c.Status(fiber.StatusOK).JSON(teamData)
}

func UpdateTeamInformation(c *fiber.Ctx) error {
	validate := validator.New()
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	// 從 URL 中獲取 secret
	secret := c.Params("secret")

	// 查詢對應的團隊
	var editSecret models.EditSecret
	err := database.Db.Collection("edit_secrets").FindOne(c.Context(), bson.M{"secret": secret}).Decode(&editSecret)
	if err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Team not found or invalid secret",
		})
	}

	// 準備解析更新的數據
	formData := new(struct {
		TeamName            string                      `json:"teamName" validate:"required,min=2,max=30"`
		TeamMembers         []models.Users              `json:"teamMembers" validate:"required,min=1,dive"`
		AccompanyingPersons []models.AccompanyingPerson `json:"accompanyingPersons" validate:"max=2,dive"`
		Exhibitors          []models.Exhibitor          `json:"exhibitors,omitempty" validate:"dive"`
	})

	// 解析請求的 JSON 數據
	if err := c.BodyParser(formData); err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// 驗證數據
	if err := validate.Struct(formData); err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 更新團隊信息
	teamID := editSecret.TeamID
	var teamData models.Team

	err = database.Db.Collection("teams").FindOne(
		c.Context(),
		bson.M{"id": teamID},
	).Decode(&teamData)

	if err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 500, "Failed to update team information", err.Error())
	}

	if len(teamData.TeamMembers) != len(formData.TeamMembers) {
		return utils.ResponseMsg(c, 400, "You can't change the number of team members", nil)
	}

	tempTeamMembers := []string{}
	// 更新團隊成員
	for i, memberID := range teamData.TeamMembers {
		var member models.Users
		err := database.Db.Collection("users").FindOne(c.Context(), bson.M{"id": memberID}).Decode(&member)
		if err != nil {
			fmt.Println(err.Error())
			return utils.ResponseMsg(c, 500, "Failed to retrieve team member", err.Error())
		}
		member.UpdatedAt = time.Now()

		for j := range member.EmergencyContacts {
			encryptedPhone := utils.EncryptAES(formData.TeamMembers[i].EmergencyContacts[j].Phone, encryptionKey)
			formData.TeamMembers[i].EmergencyContacts[j].Phone = encryptedPhone
			encryptedRelationship := utils.EncryptAES(formData.TeamMembers[i].EmergencyContacts[j].Relationship, encryptionKey)
			formData.TeamMembers[i].EmergencyContacts[j].Relationship = encryptedRelationship
			encryptedName := utils.EncryptAES(formData.TeamMembers[i].EmergencyContacts[j].Name, encryptionKey)
			formData.TeamMembers[i].EmergencyContacts[j].Name = encryptedName
		}
		encryptedStudentCardFront := utils.EncryptAES(formData.TeamMembers[i].StudentCardFront, encryptionKey)
		formData.TeamMembers[i].StudentCardFront = encryptedStudentCardFront
			
		encryptedStudentCardBack := utils.EncryptAES(formData.TeamMembers[i].StudentCardBack, encryptionKey)
		formData.TeamMembers[i].StudentCardBack = encryptedStudentCardBack

		phone := utils.EncryptAES(formData.TeamMembers[i].Phone, encryptionKey)
		formData.TeamMembers[i].Phone = phone
		birthday := utils.EncryptAES(formData.TeamMembers[i].Birthday, encryptionKey)
		formData.TeamMembers[i].Birthday = birthday
		identityNumber := utils.EncryptAES(formData.TeamMembers[i].IdentityNumber, encryptionKey)
		formData.TeamMembers[i].IdentityNumber = identityNumber

		school := utils.EncryptAES(formData.TeamMembers[i].School, encryptionKey)
		formData.TeamMembers[i].School = school
		name := utils.EncryptAES(formData.TeamMembers[i].Name, encryptionKey)
		formData.TeamMembers[i].Name = name
		remarks := utils.EncryptAES(formData.TeamMembers[i].Remarks, encryptionKey)
		formData.TeamMembers[i].Remarks = remarks
		specialDiseases := utils.EncryptAES(formData.TeamMembers[i].SpecialDiseases, encryptionKey)
		formData.TeamMembers[i].SpecialDiseases = specialDiseases
		allergies := utils.EncryptAES(formData.TeamMembers[i].Allergies, encryptionKey)
		formData.TeamMembers[i].Allergies = allergies
	
		// 檢查 Email 是否更改
		if member.Email != "" && member.Email != formData.TeamMembers[i].Email {
			randomStr, err := utils.GenerateRandomString(128)
			if err != nil {
				fmt.Println(err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to generate random string",
				})
			}

			userVerification := &models.UserVerification{
				UserID: member.ID,
				Secret: randomStr,
			}
			_, err = database.Db.Collection("user_verifications_secret").InsertOne(c.Context(), userVerification)
			if err != nil {
				fmt.Println(err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to save random string for user",
				})
			}
			// 生成驗證 URL
			baseURL := os.Getenv("BASE_URL") + "/users/verification/"
			verificationURL := fmt.Sprintf("%s%s", baseURL, randomStr)

			// 更新用戶資料
			username := formData.TeamMembers[i].Name
			_, err = database.Db.Collection("users").UpdateOne(c.Context(), bson.M{"id": memberID}, bson.M{"$set": formData.TeamMembers[i]})
			if err != nil {
				fmt.Println(err.Error())
				return utils.ResponseMsg(c, 500, "Failed to update team member", err.Error())
			}

			// 發送驗證郵件
			t, err := template.New("email_verification").Parse(htmlTempalte.VerificationTemplate)
			if err != nil {
				fmt.Println(err.Error())
				return utils.ResponseMsg(c, 500, "Failed to parse email template", err.Error())
			}

			var body bytes.Buffer
			type EmailData struct {
				Name             string
				VerificationLink string
			}
			
			if err := t.Execute(&body, EmailData{Name: username, VerificationLink: verificationURL}); err != nil {
				fmt.Println(err.Error())
				return utils.ResponseMsg(c, 500, "Failed to execute email template", err.Error())
			}

			err = utils.SendEmail(member.Email, "[Hackit] 驗證您的郵件", body.String())
			if err != nil {
				fmt.Println(err.Error())
				return utils.ResponseMsg(c, 500, fmt.Sprintf("Failed to send email to %s", member.Email), err.Error())
			}
		} else {
			// 如果 email 沒有變更，直接更新其他資料
			_, err = database.Db.Collection("users").UpdateOne(c.Context(), bson.M{"id": member.ID}, bson.M{"$set": formData.TeamMembers[i]})
			if err != nil {
				fmt.Println(err.Error())
				return utils.ResponseMsg(c, 500, "Failed to update team member", err.Error())
			}
		}
		tempTeamMembers = append(tempTeamMembers, member.ID)
	}
	teamData.TeamMembers = tempTeamMembers

	teamData.AccompanyingPersons = []string{}

	// 更新陪同人員
	for _, person := range formData.AccompanyingPersons {
		if person.ID == "" {
			person.ID = uuid.New().String() // 如果 ID 為空，創建新 UUID
		}
		phone := utils.EncryptAES(person.Phone, encryptionKey)
		person.Phone = phone
		_, err := database.Db.Collection("accompanying_persons").UpdateOne(
			c.Context(),
			bson.M{"id": person.ID},          // 根據 ID 查找
			bson.M{"$set": person},           // 更新或插入 person 資料
			options.Update().SetUpsert(true), // 設置 upsert 為 true
		)
		if err != nil {
			fmt.Println(err.Error())
			return utils.ResponseMsg(c, 500, "Failed to update accompanying person", err.Error())
		}
		teamData.AccompanyingPersons = append(teamData.AccompanyingPersons, person.ID)
	}

	teamData.Exhibitors = []string{}
	// 更新參展人
	for _, exhibitor := range formData.Exhibitors {
		if exhibitor.ID == "" {
			exhibitor.ID = uuid.New().String() // 如果 ID 為空，創建新 UUID
		}
		_, err := database.Db.Collection("exhibitors").UpdateOne(
			c.Context(),
			bson.M{"id": exhibitor.ID},       // 根據 ID 進行查找
			bson.M{"$set": exhibitor},        // 更新或插入 exhibitor 資料
			options.Update().SetUpsert(true), // 設置 upsert 為 true
		)
		if err != nil {
			fmt.Println(err.Error())
			return utils.ResponseMsg(c, 500, "Failed to update exhibitor", err.Error())
		}
		teamData.Exhibitors = append(teamData.Exhibitors, exhibitor.ID)
	}

	teamData.UpdatedAt = time.Now()
	teamData.TeamName = formData.TeamName
	err = database.Db.Collection("teams").FindOneAndUpdate(
		c.Context(),
		bson.M{"id": teamID},
		bson.M{"$set": teamData},
	).Decode(&teamData)
	if err != nil {
		fmt.Println(err.Error())
		return utils.ResponseMsg(c, 500, "Failed to update team information", err.Error())
	}

	return utils.ResponseMsg(c, 200, "Team updated successfully", nil)
}
