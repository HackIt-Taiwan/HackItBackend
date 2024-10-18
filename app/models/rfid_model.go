package models

import "time"

type VerificationRequest struct {
	UserID     string `json:"id" bson:"id"`
	UserNumber string `json:"user_number" bson:"user_number"`
	Secret     string `json:"secret" bson:"secret"`
	ValidCode  string `json:"valid_code" bson:"valid_code"`
}

type VerificationSession struct {
	UserID string `json:"id" bson:"id"`
	Jwt    string `json:"jwt" bson:"jwt"`
}

type RFIDCard struct {
	UserID     string    `json:"user_id" bson:"user_id"`
	CardNumber string    `json:"card_number" bson:"card_number"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}
