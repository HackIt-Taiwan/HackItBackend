package controllers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
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
)

var ctx = context.Background()

func CreateUsers(c *fiber.Ctx) error {
	users := new(models.Users)

	if err := c.BodyParser(users); err != nil {
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
		return utils.ResponseMsg(c, 500, "Failed to insert data", err.Error())
	}

	// TODO: Send verification email with the code

	return utils.ResponseMsg(c, 200, "User created successfully. Check email for verification code.", users)
}


func CreateTeam(c *fiber.Ctx) error {
    validate := validator.New()

    formData := new(struct {
        TeamName           string                   `json:"teamName" validate:"required,min=2,max=30"`
        TeamMembers        []models.Users           `json:"teamMembers" validate:"required,min=1,dive"`
        AccompanyingPersons []models.AccompanyingPerson `json:"accompanyingPersons" validate:"max=2,dive"`
        Exhibitors         []models.Exhibitor       `json:"exhibitors,omitempty" validate:"dive"`
    })

    // 解析請求的 JSON 數據
    if err := c.BodyParser(formData); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid input",
        })
    }

    // 驗證數據
    if err := validate.Struct(formData); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    // 將 TeamMember、AccompanyingPerson 和 Exhibitor 實例插入資料庫並獲取其 UUID
    var teamMemberIDs []string
    for _, member := range formData.TeamMembers {
		member.ID = uuid.New().String()
		member.VerificationCode = utils.GenerateVerificationCode()
		member.CheckedIn = false
		member.CreatedAt = time.Now()
		member.UpdatedAt = time.Now()
	
		// 生成128字元的隨機字符串
		randomStr, err := utils.GenerateRandomString(128)
		if err != nil {
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save random string for user",
			})
		}
	
		// 插入 team member
		_, err = database.Db.Collection("users").InsertOne(c.Context(), member)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create team member",
			})
		}
	
		teamMemberIDs = append(teamMemberIDs, member.ID)
	
		// 生成驗證 URL
		baseURL := os.Getenv("BASE_URL") + "/users/verification/"
		verificationURL := fmt.Sprintf("%s%s", baseURL, randomStr)
		
		t, err := template.New("email_verification").Parse(htmlTempalte.VerificationTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse email template: %w", err)
		}
	
		var body bytes.Buffer
		type EmailData struct {
			Name string
			VerificationLink string
		}

		if err := t.Execute(&body, EmailData{Name: member.Name, VerificationLink: verificationURL}); err != nil {
			return fmt.Errorf("failed to execute email template: %w", err)
		}

		// 發送驗證郵件
		err = utils.SendEmail(member.Email, "[Hackit] 驗證您的郵件", body.String())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to send email to %s: %v", member.Email, err),
			})
		}
	}

    var accompanyingPersonIDs []string
    for _, person := range formData.AccompanyingPersons {
        person.ID = uuid.New().String()
        _, err := database.Db.Collection("accompanying_persons").InsertOne(c.Context(), person)
        if err != nil {
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
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Failed to create exhibitor",
            })
        }
        exhibitorIDs = append(exhibitorIDs, exhibitor.ID)
    }

    // 準備將表單數據轉換為 Team 實例
    team := &models.Team{
        ID:                    uuid.New().String(),
        TeamName:              formData.TeamName,
        CreatedAt:             time.Now(),
        UpdatedAt:             time.Now(),
        TeamMembers:           teamMemberIDs,
        AccompanyingPersons:   accompanyingPersonIDs,
        Exhibitors:            exhibitorIDs,
    }

    // 插入 Team 到 MongoDB
    collection := database.Db.Collection("teams")
    _, err := collection.InsertOne(c.Context(), team)
    if err != nil {
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
		if err == mongo.ErrNoDocuments {
			return utils.ResponseMsg(c, 404, "Verification link is invalid or has expired", nil)
		}
		return utils.ResponseMsg(c, 500, "Error checking verification secret", err.Error())
	}

	// 根據找到的 userID 更新使用者的驗證狀態
	filter := bson.M{"id": userVerification.UserID}
	update := bson.M{"$set": bson.M{"verified": true}}
	_, err = database.Db.Collection("users").UpdateOne(c.Context(), filter, update)
	if err != nil {
		return utils.ResponseMsg(c, 500, "Failed to update user verification status", err.Error())
	}
	
	_, err = database.Db.Collection("user_verifications_secret").DeleteOne(c.Context(), bson.M{"secret": secret})
	if err != nil {
		return utils.ResponseMsg(c, 500, "Failed to delete verification token", err.Error())
	}

	return c.Redirect(os.Getenv("BASE_URL_FRONTEND") + "/verification/success", fiber.StatusSeeOther)
}