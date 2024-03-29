package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

// Token structure object
type Token struct {
	UserID uint
	Name   string
	Email  string
	Role   int
	*jwt.StandardClaims
}

// Handler handler struct
type Handler struct {
	DB *gorm.DB
}

// Credential object struct
type Credential struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims struct
type Claims struct {
	jwt.StandardClaims
}

// TribeLeadAssign user lead
type TribeLeadAssign struct {
	LeadID  uint `gorm:"primary_key"`
	TribeID uint `gorm:"primary_key"`
}