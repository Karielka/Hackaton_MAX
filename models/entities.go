package models

import "gorm.io/gorm"

// Базовые сущности под твои экраны (минимальные поля для старта)

type University struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"index;not null"`
}

type Faculty struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"index;not null"`
	UniversityID uint
}

type Teacher struct {
	ID        uint   `gorm:"primaryKey"`
	FullName  string `gorm:"index;not null"`
	Email     string
	FacultyID uint
	Subject   string // предмет/кафедра
	// расписание можно хранить в JSON в отдельной таблице — пока опускаем
}

type DeanOffice struct {
	ID         uint   `gorm:"primaryKey"`
	FacultyID  uint   `gorm:"uniqueIndex"`
	Schedule   string // текст/линк – заглушка
	DocsLink   string // ссылка на документы
	Contacts   string // email/телефон
}

type Campus struct {
	ID           uint   `gorm:"primaryKey"`
	UniversityID uint
	ShortName    string `gorm:"index"`
	Address      string
	ImageURL     string
}

type Place struct { // столовая/буфет/копирка и т.д.
	ID       uint   `gorm:"primaryKey"`
	CampusID uint   `gorm:"index"`
	Type     string `gorm:"index"` // "canteen" | "buffet" | "copy"
	Name     string
	Location string
	Schedule string
	MenuURL  string // ссылка на меню на сегодня (если есть)
}

type FAQ struct {
	ID       uint   `gorm:"primaryKey"`
	Question string `gorm:"index"`
	Answer   string
}

// AutoMigrate выполняем из main
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&University{},
		&Faculty{},
		&Teacher{},
		&DeanOffice{},
		&Campus{},
		&Place{},
		&FAQ{},
	)
}
