package password

import (
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "отрицательная длина",
			config:  Config{Length: -1, UseDigits: true},
			wantErr: true,
		},
		{
			name:    "нулевая длина",
			config:  Config{Length: 0, UseDigits: true},
			wantErr: true,
		},
		{
			name:    "нет выбранных наборов",
			config:  Config{Length: 10, UseDigits: false, UseLower: false, UseUpper: false},
			wantErr: true,
		},
		{
			name:    "валидная конфигурация - только digits",
			config:  Config{Length: 5, UseDigits: true},
			wantErr: false,
		},
		{
			name:    "валидная конфигурация - все наборы",
			config:  Config{Length: 10, UseDigits: true, UseLower: true, UseUpper: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildCharset(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		wantLen      int
		wantCharsets int
	}{
		{
			name:         "только digits",
			config:       Config{UseDigits: true},
			wantLen:      10,
			wantCharsets: 1,
		},
		{
			name:         "только lower",
			config:       Config{UseLower: true},
			wantLen:      26,
			wantCharsets: 1,
		},
		{
			name:         "только upper",
			config:       Config{UseUpper: true},
			wantLen:      26,
			wantCharsets: 1,
		},
		{
			name:         "все вместе",
			config:       Config{UseDigits: true, UseLower: true, UseUpper: true},
			wantLen:      62,
			wantCharsets: 3,
		},
		{
			name:         "digits и lower",
			config:       Config{UseDigits: true, UseLower: true},
			wantLen:      36,
			wantCharsets: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			charset, charsets := buildCharset(tt.config)
			if len(charset) != tt.wantLen {
				t.Errorf("buildCharset() charset length = %v, want %v", len(charset), tt.wantLen)
			}
			if len(charsets) != tt.wantCharsets {
				t.Errorf("buildCharset() charsets count = %v, want %v", len(charsets), tt.wantCharsets)
			}
		})
	}
}

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "длина больше charset",
			config:  Config{Length: 11, UseDigits: true}, // digits только 10
			wantErr: true,
		},
		{
			name:    "валидная конфигурация",
			config:  Config{Length: 10, UseDigits: true, UseLower: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && gen == nil {
				t.Errorf("NewGenerator() returned nil generator")
			}
		})
	}
}

func TestGenerateSinglePassword(t *testing.T) {
	config := Config{
		Length:    12,
		UseDigits: true,
		UseLower:  true,
		UseUpper:  true,
	}

	gen, err := NewGenerator(config)
	if err != nil {
		t.Fatalf("NewGenerator() failed: %v", err)
	}

	password, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Проверяем длину
	if len(password) != config.Length {
		t.Errorf("Password length = %d, want %d", len(password), config.Length)
	}

	// Проверяем отсутствие повторов
	seen := make(map[rune]bool)
	for _, char := range password {
		if seen[char] {
			t.Errorf("Password contains duplicate character: %c", char)
		}
		seen[char] = true
	}

	// Проверяем, что все символы из разрешённого набора
	validChars := digits + lower + upper
	for _, char := range password {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("Password contains invalid character: %c", char)
		}
	}
}

func TestGenerateWithAllCharsets(t *testing.T) {
	config := Config{
		Length:    15,
		UseDigits: true,
		UseLower:  true,
		UseUpper:  true,
	}

	gen, err := NewGenerator(config)
	if err != nil {
		t.Fatalf("NewGenerator() failed: %v", err)
	}

	// Генерируем несколько паролей для уверенности
	for i := 0; i < 10; i++ {
		password, err := gen.Generate()
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		hasDigit := false
		hasLower := false
		hasUpper := false

		for _, char := range password {
			if strings.ContainsRune(digits, char) {
				hasDigit = true
			}
			if strings.ContainsRune(lower, char) {
				hasLower = true
			}
			if strings.ContainsRune(upper, char) {
				hasUpper = true
			}
		}

		if !hasDigit {
			t.Errorf("Password %q doesn't contain any digit", password)
		}
		if !hasLower {
			t.Errorf("Password %q doesn't contain any lowercase letter", password)
		}
		if !hasUpper {
			t.Errorf("Password %q doesn't contain any uppercase letter", password)
		}
	}
}

func TestGenerateUnique(t *testing.T) {
	config := Config{
		Length:    10,
		UseDigits: true,
		UseLower:  true,
		UseUpper:  true,
	}

	gen, err := NewGenerator(config)
	if err != nil {
		t.Fatalf("NewGenerator() failed: %v", err)
	}

	count := 50
	passwords, err := gen.GenerateUnique(count)
	if err != nil {
		t.Fatalf("GenerateUnique() failed: %v", err)
	}

	if len(passwords) != count {
		t.Errorf("GenerateUnique() returned %d passwords, want %d", len(passwords), count)
	}

	// Проверяем уникальность через map
	seen := make(map[string]bool)
	for _, pwd := range passwords {
		if seen[pwd] {
			t.Errorf("Duplicate password found: %s", pwd)
		}
		seen[pwd] = true
	}
}

func TestGenerateMaxPossiblePasswords(t *testing.T) {
	// Малый charset: только digits (10 символов), длина 5
	// Максимум: P(10, 5) = 30240 комбинаций
	config := Config{
		Length:    5,
		UseDigits: true,
	}

	gen, err := NewGenerator(config)
	if err != nil {
		t.Fatalf("NewGenerator() failed: %v", err)
	}

	// Генерируем разумное количество
	count := 100
	passwords, err := gen.GenerateUnique(count)
	if err != nil {
		t.Fatalf("GenerateUnique() failed: %v", err)
	}

	// Проверяем уникальность
	seen := make(map[string]bool)
	for _, pwd := range passwords {
		if seen[pwd] {
			t.Errorf("Duplicate password found: %s", pwd)
		}
		seen[pwd] = true
	}
}

func TestGenerateExceedsLimit(t *testing.T) {
	// Очень маленький charset и большой count
	config := Config{
		Length:    3,
		UseDigits: true, // 10 символов
	}

	gen, err := NewGenerator(config)
	if err != nil {
		t.Fatalf("NewGenerator() failed: %v", err)
	}

	// P(10, 3) = 720 возможных комбинаций
	// Если запросим больше и с учётом правила уникальности, в какой-то момент должна быть ошибка
	// Но для теста просто проверим, что при попытке генерации огромного числа паролей
	// с малым charset рано или поздно получим ошибку

	// Установим меньший maxAttempts для теста
	gen.maxAttempts = 100

	// Генерируем много паролей (больше чем возможно при низком лимите попыток)
	_, err = gen.GenerateUnique(1000)
	if err == nil {
		t.Error("Expected error when generating too many passwords, got none")
	}
}

func TestGenerateUniqueInvalidCount(t *testing.T) {
	config := Config{
		Length:    10,
		UseDigits: true,
		UseLower:  true,
	}

	gen, err := NewGenerator(config)
	if err != nil {
		t.Fatalf("NewGenerator() failed: %v", err)
	}

	_, err = gen.GenerateUnique(0)
	if err == nil {
		t.Error("Expected error for count=0, got none")
	}

	_, err = gen.GenerateUnique(-5)
	if err == nil {
		t.Error("Expected error for negative count, got none")
	}
}

func TestNoRepeatedCharactersInPassword(t *testing.T) {
	config := Config{
		Length:    20,
		UseDigits: true,
		UseLower:  true,
		UseUpper:  true,
	}

	gen, err := NewGenerator(config)
	if err != nil {
		t.Fatalf("NewGenerator() failed: %v", err)
	}

	passwords, err := gen.GenerateUnique(100)
	if err != nil {
		t.Fatalf("GenerateUnique() failed: %v", err)
	}

	for _, password := range passwords {
		charCount := make(map[rune]int)
		for _, char := range password {
			charCount[char]++
			if charCount[char] > 1 {
				t.Errorf("Password %q has repeated character %c", password, char)
			}
		}
	}
}
