package models

import (
	"time"
)

type Team struct {
	ID                  string    `json:"id,omitempty" bson:"id,omitempty"` // 使用 id 作為 BSON 標記
	TeamName            string    `json:"teamName,omitempty" bson:"teamName,omitempty"`
	CreatedAt           time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	TeamMembers         []string  `json:"teamMembers,omitempty" bson:"teamMembers,omitempty"`                 // 只存 UUID
	AccompanyingPersons []string  `json:"accompanyingPersons,omitempty" bson:"accompanyingPersons,omitempty"` // 只存 UUID
	Exhibitors          []string  `json:"exhibitors,omitempty" bson:"exhibitors,omitempty"`                   // 只存 UUID
}

// TeamMember 模型
type Users struct {
	ID                string             `json:"id,omitempty" bson:"id,omitempty"` // 使用 id 作為 BSON 標記
	IsRepresentative  bool               `json:"isRepresentative,omitempty" bson:"isRepresentative,omitempty"`
	Name              string             `json:"name,omitempty" bson:"name,omitempty"`
	Gender            string             `json:"gender,omitempty" bson:"gender,omitempty"`
	School            string             `json:"school,omitempty" bson:"school,omitempty"`
	Grade             string             `json:"grade,omitempty" bson:"grade,omitempty"`
	IdentityNumber    string             `json:"identityNumber,omitempty" bson:"identityNumber,omitempty"`
	StudentCardFront  string             `json:"studentCardFront,omitempty" bson:"studentCardFront,omitempty"`
	StudentCardBack   string             `json:"studentCardBack,omitempty" bson:"studentCardBack,omitempty"`
	UserNumber        int                `json:"userNumber,omitempty" bson:"userNumber,omitempty"`
	Birthday          string             `json:"birthday,omitempty" bson:"birthday,omitempty"`
	Email             string             `json:"email,omitempty" bson:"email,omitempty"`
	Phone             string             `json:"phone,omitempty" bson:"phone,omitempty"`
	EmergencyContacts []EmergencyContact `json:"emergencyContacts,omitempty" bson:"emergencyContacts,omitempty"`
	TShirtSize        string             `json:"tShirtSize,omitempty" bson:"tShirtSize,omitempty"`
	Allergies         string             `json:"allergies,omitempty" bson:"allergies,omitempty"`
	SpecialDiseases   string             `json:"specialDiseases,omitempty" bson:"specialDiseases,omitempty"`
	Remarks           string             `json:"remarks,omitempty" bson:"remarks,omitempty"`
	VerificationCode  string             `json:"verificationCode,omitempty" bson:"verificationCode,omitempty"`
	CheckedIn         bool               `json:"checkedIn,omitempty" bson:"checkedIn,omitempty"`
	CreatedAt         time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty" binding:"required"`
	UpdatedAt         time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty" binding:"required"`
}

// EmergencyContact 模型
type EmergencyContact struct {
	Name         string `json:"name,omitempty" bson:"name,omitempty"`
	Relationship string `json:"relationship,omitempty" bson:"relationship,omitempty"`
	Phone        string `json:"phone,omitempty" bson:"phone,omitempty"`
}

// AccompanyingPerson 模型
type AccompanyingPerson struct {
	ID    string `json:"id,omitempty" bson:"id,omitempty"` // 使用 id 作為 BSON 標記
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
	Phone string `json:"phone,omitempty" bson:"phone,omitempty"`
}

// Exhibitor 模型
type Exhibitor struct {
	ID    string `json:"id,omitempty" bson:"id,omitempty"` // 使用 id 作為 BSON 標記
	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
}

type UserVerification struct {
	UserID string `json:"user_id" bson:"user_id"`
	Secret string `json:"secret" bson:"secret"`
}

type EditSecret struct {
	TeamID string `json:"team_id" bson:"team_id"`
	Secret string `json:"secret" bson:"secret"`
}
