package models

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// SeedSampleData наполняет БД тестовыми данными:
// Institute -> Faculty("ИУ") -> Departments("ИУ1".."ИУ10") -> Teachers(2-3 на кафедру)
func SeedSampleData(db *gorm.DB) error {
	// 1) Институт (общий)
	inst := Institute{Name: "Институт информатики и систем управления"}
	if err := db.Where("name = ?", inst.Name).FirstOrCreate(&inst).Error; err != nil {
		return fmt.Errorf("seed institute: %w", err)
	}
	// Добавляем корпуса
	campuses := []Campus{
		{
			ShortName:   "ГУК",
			FullName:    "Главный учебный корпус",
			Address:     "ул. Студенческая, д. 35",
			Metro:       "«Университетская» (5 мин. пешком)",
			ImageURL:    "https://example.com/images/guk.jpg",
			MapImageURL: "https://example.com/maps/guk_map.jpg",
			Description: "• Аудитории 100-499\n• Деканат ФМиЕН\n• Столовая №1\n• Библиотека",
		},
		{
			ShortName:   "Корпус 2",
			FullName:    "Второй учебный корпус",
			Address:     "ул. Академическая, д. 15",
			Metro:       "«Научная» (10 мин. пешком)",
			ImageURL:    "https://example.com/images/corpus2.jpg",
			MapImageURL: "https://example.com/maps/corpus2_map.jpg",
			Description: "• Аудитории 500-799\n• Лаборатории физики\n• Буфет №2\n• Спортивный зал",
		},
	}

	for _, campus := range campuses {
		if err := db.FirstOrCreate(&campus, Campus{ShortName: campus.ShortName}).Error; err != nil {
			return err
		}
	}

	// Добавляем места (столовые, буфеты, копирки)
	places := []Place{
		// Столовые в ГУК
		{
			CampusID:  1, // ГУК
			Type:      "canteen",
			Name:      "Столовая №1",
			Location:  "1 этаж, левое крыло",
			Schedule:  "Пн-Пт 9:00 - 17:00",
			MenuToday: "• Комплексный обед №1 (120 руб): Борщ, пюре с котлетой\n• Комплексный обед №2 (110 руб): Салат, гречка с курицей\n• Плов (90 руб)\n• Салат \"Цезарь\" (70 руб)",
		},
		// Буфеты в ГУК
		{
			CampusID:  1, // ГУК
			Type:      "buffet",
			Name:      "Буфет №1",
			Location:  "2 этаж, центральный холл",
			Schedule:  "Пн-Пт 8:00 - 20:00",
			MenuToday: "• Выпечка (30-50 руб)\n• Сэндвичи (60-80 руб)\n• Напитки (40-60 руб)\n• Снеки (20-40 руб)",
		},
		{
			CampusID:  1, // ГУК
			Type:      "buffet",
			Name:      "Буфет №2",
			Location:  "3 этаж, правое крыло",
			Schedule:  "Пн-Пт 9:00 - 18:00",
			MenuToday: "• Кофе (50-80 руб)\n• Чай (30 руб)\n• Печенье (25 руб)\n• Шоколад (40 руб)",
		},
		// Копирки в ГУК
		{
			CampusID:  1, // ГУК
			Type:      "copy",
			Name:      "Копировальный центр №1",
			Location:  "1 этаж, рядом с библиотекой",
			Schedule:  "Пн-Пт 9:00 - 18:00",
			MenuToday: "• Печать ч/б (5 руб/стр)\n• Печать цветная (20 руб/стр)\n• Ксерокопия (3 руб/стр)\n• Брошюровка (от 50 руб)",
		},

		// Места в Корпусе 2
		{
			CampusID:  2, // Корпус 2
			Type:      "canteen",
			Name:      "Столовая №2",
			Location:  "Цокольный этаж",
			Schedule:  "Пн-Пт 10:00 - 16:00",
			MenuToday: "• Комплексный обед (100 руб): Суп, гарнир с мясом\n• Салаты (40-60 руб)\n• Напитки (30-50 руб)",
		},
		{
			CampusID:  2, // Корпус 2
			Type:      "copy",
			Name:      "Копировальный центр №2",
			Location:  "2 этаж, каб. 215",
			Schedule:  "Пн-Пт 10:00 - 17:00",
			MenuToday: "• Печать ч/б (4 руб/стр)\n• Сканирование (10 руб/стр)\n• Ламинирование (30 руб/стр)",
		},
	}

	for _, place := range places {
		if err := db.FirstOrCreate(&place,
			Place{CampusID: place.CampusID, Type: place.Type, Name: place.Name}).Error; err != nil {
			return err
		}
	}

	// 2) Факультет "ИУ"
	faculties := []Faculty{
		{Name: "ИУ", InstituteID: inst.ID},
		{Name: "Э", InstituteID: inst.ID},
		{Name: "РК", InstituteID: inst.ID},
	}
	for i, f := range faculties {
		if err := db.Where("name = ?", f.Name).FirstOrCreate(&faculties[i]).Error; err != nil {
			return fmt.Errorf("seed faculty %s: %w", f.Name, err)
		}
	}

	// 3) Кафедры ИУ1..ИУ10
	for _, fac := range faculties {
		for i := 1; i <= 5; i++ {
			depName := fmt.Sprintf("%s%d", fac.Name, i)
			dep := Department{Name: depName, FacultyID: fac.ID}
			if err := db.Where("name = ?", dep.Name).FirstOrCreate(&dep).Error; err != nil {
				return fmt.Errorf("seed department %s: %w", depName, err)
			}

			teachers := sampleTeachersFor(depName, dep.ID, i)
			for _, t := range teachers {
				var existing Teacher
				err := db.Where("full_name = ? AND department_id = ?", t.FullName, dep.ID).First(&existing).Error
				if err == gorm.ErrRecordNotFound {
					if err := db.Create(&t).Error; err != nil {
						return fmt.Errorf("seed teacher %s/%s: %w", depName, t.FullName, err)
					}
				} else if err != nil {
					return fmt.Errorf("check teacher %s/%s: %w", depName, t.FullName, err)
				}
			}
		}
	}

	// --- 5) Деканаты (DeanOffice) для каждого факультета
	for _, fac := range faculties {
		office := DeanOffice{
			FacultyID: fac.ID,
			Schedule: fmt.Sprintf(
				"Пн–Чт: 10:00–17:00 (обед 13:00–14:00)\nПт: 10:00–16:00\nСб–Вс: выходной\n\nОтветственный секретарь: %s",
				randomSecretary(fac.Name),
			),
			DocsLink: fmt.Sprintf("https://example.edu/%s/dean/docs", strings.ToLower(fac.Name)),
			Contacts: fmt.Sprintf("Тел.: +7 (495) 000-00-%03d, каб. %s-204", fac.ID+100, fac.Name),
		}
		if err := db.Where("faculty_id = ?", office.FacultyID).
			Assign(office). // если перезапускать сид, обновим данные
			FirstOrCreate(&DeanOffice{}).Error; err != nil {
			return fmt.Errorf("seed dean office for faculty %s: %w", fac.Name, err)
		}
	}

	return nil
}


// sampleTeachersFor возвращает 3 преподавателя с разными ФИО
func sampleTeachersFor(depName string, depID uint, depIndex int) []Teacher {
	firstNames := []string{"Иван", "Пётр", "Анна", "Екатерина", "Сергей", "Мария", "Дмитрий", "Ольга", "Алексей", "Наталья"}
	lastNames := []string{"Иванов", "Петров", "Сидорова", "Кузнецов", "Смирнова", "Попов", "Лебедев", "Козлова", "Новикова", "Морозов"}
	middles := []string{"Иванович", "Петрович", "Сергеевна", "Андреевна", "Алексеевич", "Владимировна"}

	// helper для кругового выбора из массивов
	pick := func(arr []string, i int) string { return arr[i%len(arr)] }

	makeT := func(i int, subj, days, room string) Teacher {
		fn := pick(firstNames, depIndex+i)
		ln := pick(lastNames, depIndex*2+i)
		mn := pick(middles, depIndex+i/2)
		full := fmt.Sprintf("%s %s %s", ln, fn, mn)

		return Teacher{
			FullName:     full,
			Email:        fmt.Sprintf("%s_%s@example.edu", translit(depName), strings.ToLower(ln)),
			Subject:      subj,
			DepartmentID: depID,
			Schedule:     fmt.Sprintf("%s; Аудитория: %s", days, room),
		}
	}

	return []Teacher{
		makeT(0, "Алгоритмы и структуры данных", "Пн 10:00–11:40; Ср 12:00–13:40", "А-101"),
		makeT(1, "Базы данных", "Вт 14:00–15:40; Чт 10:00–11:40", "Б-203"),
		makeT(2, "Операционные системы", "Пт 09:00–10:40; Ср 16:00–17:40", "В-317"),
	}
}

// translit — упрощённая замена для email-части
func translit(dep string) string {
	repl := map[rune]string{
		'И': "iu", 'и': "iu",
		'У': "u", 'у': "u",
		' ': "", '-': "-",
		'0': "0", '1': "1", '2': "2", '3': "3", '4': "4",
		'5': "5", '6': "6", '7': "7", '8': "8", '9': "9",
	}
	out := make([]rune, 0, len(dep))
	for _, r := range dep {
		if v, ok := repl[r]; ok {
			for _, rr := range v {
				out = append(out, rr)
			}
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

// randomSecretary возвращает демонстрационного ответственного секретаря
func randomSecretary(fac string) string {
	switch fac {
	case "ИУ":
		return "Иванова Елена Сергеевна"
	case "Э":
		return "Петров Алексей Владимирович"
	case "РК":
		return "Сидорова Наталья Ивановна"
	default:
		return "Неизвестен"
	}
}
