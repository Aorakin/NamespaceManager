package dtos

type RegisterInput struct {
	Username string `json:"username" gorm:"not null" validate:"required,min=2,max=30"`
	Password string `json:"password" gorm:"not null" validate:"required,min=8"`
	Gmail    string `json:"email" gorm:"not null" validate:"required,email"`
}

type LoginInput struct {
	Username string `json:"username" gorm:"not null"`
	Password string `json:"password" gorm:"not null"`
}
