package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vikto/passwordgen/internal/password"
)

func main() {
	// Определяем флаги
	var (
		length  int
		lengthL int
		digits  bool
		lower   bool
		upper   bool
		count   int
	)

	flag.IntVar(&length, "length", 0, "Длина пароля (обязательный параметр)")
	flag.IntVar(&lengthL, "l", 0, "Длина пароля (короткий вариант)")
	flag.BoolVar(&digits, "digits", false, "Использовать цифры 0-9")
	flag.BoolVar(&lower, "lower", false, "Использовать маленькие буквы a-z")
	flag.BoolVar(&upper, "upper", false, "Использовать большие буквы A-Z")
	flag.IntVar(&count, "count", 1, "Количество паролей для генерации")

	// Кастомизируем help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Генератор уникальных паролей\n\n")
		fmt.Fprintf(os.Stderr, "Использование:\n")
		fmt.Fprintf(os.Stderr, "  %s [опции]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Примеры:\n")
		fmt.Fprintf(os.Stderr, "  %s -length 12 -digits -lower -upper\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -l 10 -digits -lower -count 5\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -length 8 -upper -count 3\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Опции:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Выбираем длину (приоритет у -length, если оба не указаны - ошибка)
	finalLength := length
	if finalLength == 0 {
		finalLength = lengthL
	}

	if finalLength <= 0 {
		fmt.Fprintf(os.Stderr, "Ошибка: необходимо указать длину пароля через -length или -l\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Проверяем, что выбран хотя бы один набор символов
	if !digits && !lower && !upper {
		fmt.Fprintf(os.Stderr, "Ошибка: необходимо выбрать хотя бы один набор символов (-digits, -lower или -upper)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Создаём конфигурацию
	config := password.Config{
		Length:    finalLength,
		UseDigits: digits,
		UseLower:  lower,
		UseUpper:  upper,
	}

	// Создаём генератор
	gen, err := password.NewGenerator(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания генератора: %v\n", err)
		os.Exit(1)
	}

	// Генерируем пароли
	passwords, err := gen.GenerateUnique(count)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка генерации паролей: %v\n", err)
		os.Exit(1)
	}

	// Выводим результат
	for _, pwd := range passwords {
		fmt.Println(pwd)
	}
}
