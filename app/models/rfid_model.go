package models

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