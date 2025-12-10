package password

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Config содержит параметры для генерации пароля
type Config struct {
	Length    int
	UseDigits bool
	UseLower  bool
	UseUpper  bool
}

// Generator генерирует уникальные пароли
type Generator struct {
	charset     []rune
	charsets    [][]rune
	length      int
	used        map[string]struct{}
	maxAttempts int
}

const (
	digits = "0123456789"
	lower  = "abcdefghijklmnopqrstuvwxyz"
	upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// NewGenerator создаёт новый генератор паролей с валидацией конфигурации
func NewGenerator(config Config) (*Generator, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	charset, charsets := buildCharset(config)

	if config.Length > len(charset) {
		return nil, fmt.Errorf("длина пароля (%d) превышает количество доступных уникальных символов (%d)", config.Length, len(charset))
	}

	return &Generator{
		charset:     charset,
		charsets:    charsets,
		length:      config.Length,
		used:        make(map[string]struct{}),
		maxAttempts: 10000, // разумный лимит попыток
	}, nil
}

// validateConfig проверяет корректность конфигурации
func validateConfig(config Config) error {
	if config.Length <= 0 {
		return fmt.Errorf("длина пароля должна быть положительным числом")
	}

	if !config.UseDigits && !config.UseLower && !config.UseUpper {
		return fmt.Errorf("необходимо выбрать хотя бы один набор символов (digits, lower или upper)")
	}

	return nil
}

// buildCharset создаёт общий набор символов и группы для валидации
func buildCharset(config Config) ([]rune, [][]rune) {
	var charset []rune
	var charsets [][]rune

	if config.UseDigits {
		digitsRunes := []rune(digits)
		charset = append(charset, digitsRunes...)
		charsets = append(charsets, digitsRunes)
	}

	if config.UseLower {
		lowerRunes := []rune(lower)
		charset = append(charset, lowerRunes...)
		charsets = append(charsets, lowerRunes)
	}

	if config.UseUpper {
		upperRunes := []rune(upper)
		charset = append(charset, upperRunes...)
		charsets = append(charsets, upperRunes)
	}

	return charset, charsets
}

// Generate генерирует один уникальный пароль
func (g *Generator) Generate() (string, error) {
	for attempt := 0; attempt < g.maxAttempts; attempt++ {
		password, err := g.generateOne()
		if err != nil {
			return "", err
		}

		// Проверяем уникальность
		if _, exists := g.used[password]; !exists {
			g.used[password] = struct{}{}
			return password, nil
		}
	}

	return "", fmt.Errorf("не удалось сгенерировать уникальный пароль за %d попыток, возможно достигнут лимит комбинаций", g.maxAttempts)
}

// generateOne генерирует один пароль (без проверки уникальности)
func (g *Generator) generateOne() (string, error) {
	// Создаём временную копию доступных символов
	available := make([]rune, len(g.charset))
	copy(available, g.charset)

	var result []rune

	// Если используется несколько наборов, гарантируем минимум один символ из каждого
	if len(g.charsets) > 1 {
		for _, charsetGroup := range g.charsets {
			// Находим символы из этой группы, которые ещё доступны
			var availableFromGroup []int
			for i, char := range available {
				if containsRune(charsetGroup, char) {
					availableFromGroup = append(availableFromGroup, i)
				}
			}

			if len(availableFromGroup) == 0 {
				return "", fmt.Errorf("недостаточно символов для удовлетворения требований")
			}

			// Выбираем случайный символ из этой группы
			randIdx, err := secureRandomInt(len(availableFromGroup))
			if err != nil {
				return "", err
			}

			selectedIdx := availableFromGroup[randIdx]
			result = append(result, available[selectedIdx])

			// Удаляем выбранный символ из available
			available = removeAtIndex(available, selectedIdx)
		}
	}

	// Заполняем оставшиеся позиции
	remaining := g.length - len(result)
	for i := 0; i < remaining; i++ {
		if len(available) == 0 {
			return "", fmt.Errorf("недостаточно уникальных символов")
		}

		randIdx, err := secureRandomInt(len(available))
		if err != nil {
			return "", err
		}

		result = append(result, available[randIdx])
		available = removeAtIndex(available, randIdx)
	}

	// Перемешиваем результат
	if err := shuffle(result); err != nil {
		return "", err
	}

	return string(result), nil
}

// GenerateUnique генерирует count уникальных паролей
func (g *Generator) GenerateUnique(count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("количество паролей должно быть положительным числом")
	}

	var result []string

	for i := 0; i < count; i++ {
		password, err := g.Generate()
		if err != nil {
			return nil, fmt.Errorf("не удалось сгенерировать %d уникальных паролей: %w", count, err)
		}
		result = append(result, password)
	}

	return result, nil
}

// secureRandomInt генерирует безопасное случайное число в диапазоне [0, max)
func secureRandomInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("максимум должен быть положительным числом")
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, fmt.Errorf("ошибка генерации случайного числа: %w", err)
	}

	return int(nBig.Int64()), nil
}

// shuffle перемешивает срез с использованием алгоритма Fisher-Yates и crypto/rand
func shuffle(slice []rune) error {
	for i := len(slice) - 1; i > 0; i-- {
		j, err := secureRandomInt(i + 1)
		if err != nil {
			return err
		}
		slice[i], slice[j] = slice[j], slice[i]
	}
	return nil
}

// removeAtIndex удаляет элемент по индексу из среза
func removeAtIndex(slice []rune, index int) []rune {
	return append(slice[:index], slice[index+1:]...)
}

// containsRune проверяет, содержит ли срез заданную руну
func containsRune(slice []rune, target rune) bool {
	for _, r := range slice {
		if r == target {
			return true
		}
	}
	return false
}
