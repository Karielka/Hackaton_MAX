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

	// 2) Факультет "ИУ"
	fac := Faculty{Name: "ИУ", InstituteID: inst.ID}
	if err := db.Where("name = ?", fac.Name).FirstOrCreate(&fac).Error; err != nil {
		return fmt.Errorf("seed faculty: %w", err)
	}

	// 3) Кафедры ИУ1..ИУ10
	for i := 1; i <= 10; i++ {
		depName := fmt.Sprintf("ИУ%d", i)
		dep := Department{Name: depName, FacultyID: fac.ID}
		if err := db.Where("name = ?", dep.Name).FirstOrCreate(&dep).Error; err != nil {
			return fmt.Errorf("seed department %s: %w", depName, err)
		}

		teachers := sampleTeachersFor(depName, dep.ID, i) // ← передаём i
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


	return nil
}

// sampleTeachersFor возвращает 2–3 демонстрационных преподавателя для кафедры
// sampleTeachersFor возвращает 3 преподавателя с разными ФИО
func sampleTeachersFor(depName string, depID uint, depIndex int) []Teacher {
	firstNames := []string{"Иван", "Пётр", "Анна", "Екатерина", "Сергей", "Мария", "Дмитрий", "Ольга", "Алексей", "Наталья"}
	lastNames  := []string{"Иванов", "Петров", "Сидорова", "Кузнецов", "Смирнова", "Попов", "Лебедев", "Козлова", "Новикова", "Морозов"}
	middles    := []string{"Иванович", "Петрович", "Сергеевна", "Андреевна", "Алексеевич", "Владимировна"}

	// helper для кругового выбора из массивов
	pick := func(arr []string, i int) string { return arr[i%len(arr)] }

	makeT := func(i int, subj, days, room string) Teacher {
		fn := pick(firstNames, depIndex+i)
		ln := pick(lastNames,  depIndex*2+i) // другое смещение → меньше повторов
		mn := pick(middles,    depIndex+i/2)
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
		makeT(1, "Базы данных",                  "Вт 14:00–15:40; Чт 10:00–11:40", "Б-203"),
		makeT(2, "Операционные системы",         "Пт 09:00–10:40; Ср 16:00–17:40", "В-317"),
	}
}


// translit — упрощённая замена для email-части
func translit(dep string) string {
	repl := map[rune]string{
		'И': "iu", 'и': "iu",
		'У': "u",  'у': "u",
		' ': "",   '-': "-",
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
