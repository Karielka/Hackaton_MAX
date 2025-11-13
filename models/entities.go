package models

import "gorm.io/gorm"

type Institute struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;not null"`
}

type Faculty struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex;not null"`
	InstituteID uint
	Institute   Institute `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Department struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex;not null"`
	FacultyID uint
	Faculty   Faculty `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Teacher struct {
	ID           uint   `gorm:"primaryKey"`
	FullName     string `gorm:"index;not null"`
	Email        string `gorm:"index"`
	Subject      string `gorm:"index"` // предмет (опционально)
	DepartmentID uint
	Department   Department `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Schedule     string     `gorm:"type:text"` // заглушка; позже вынесем в отдельную сущность
}

type DeanOffice struct {
	ID        uint `gorm:"primaryKey"`
	FacultyID uint `gorm:"uniqueIndex"`
	Schedule  string
	DocsLink  string
	Contacts  string
}

type Campus struct {
	ID          uint   `gorm:"primaryKey"`
	ShortName   string `gorm:"index;not null"`
	FullName    string `gorm:"index;not null"`
	Address     string
	Metro       string // Ближайшее метро
	ImageURL    string // Фото корпуса
	MapImageURL string // Фото с картой
	Description string // Что находится внутри
	InstituteID uint   // если нужно привязать корпуса к институту
}

type Place struct {
	ID        uint   `gorm:"primaryKey"`
	CampusID  uint   `gorm:"index"`
	Type      string `gorm:"index"` // "canteen" | "buffet" | "copy"...
	Name      string
	Location  string
	Schedule  string
	MenuURL   string
	MenuToday string `gorm:"type:text"` // Меню на сегодня
}

type FAQ struct {
	ID       uint   `gorm:"primaryKey"`
	Question string `gorm:"index"`
	Answer   string
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Institute{},
		&Faculty{},
		&Department{},
		&Teacher{},
		&DeanOffice{},
		&Campus{},
		&Place{},
		&FAQ{},
	)
}
