package models

import (
	"time"
)

type Users struct {
	ID               string    `json:"id,omitempty" bson:"id,omitempty"`
	UserNumber       string    `json:"user_number,omitempty" bson:"user_number,omitempty"`
	Username         string    `json:"username,omitempty" bson:"username,omitempty" binding:"required"`
	Email            string    `json:"email,omitempty" bson:"email,omitempty" binding:"required"`
	VerificationCode string    `json:"verificationCode,omitempty" bson:"verificationCode,omitempty"`
	IsVerified       bool      `json:"isVerified,omitempty" bson:"isVerified,omitempty"`
	CreatedAt        time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty" binding:"required"`
	UpdatedAt        time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty" binding:"required"`
}
