package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const base = "http://localhost:8082"

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Предупреждение: файл .env не найден")
	}

	token := os.Getenv("MIDDLEWARE_TOKEN")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("*    Сокращатель ссылок    *")
	fmt.Println("Введите 'exit' для выхода")

	for {
		fmt.Println("\n1. Создать ссылку")
		fmt.Println("2. Удалить ссылку")
		fmt.Println()
		fmt.Print("Введите номер действия: ")
		scanner.Scan()
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			createURL(scanner)
		case "2":
			deleteURL(scanner, token)
		case "exit":
			return
		default:
			fmt.Println("Неверный выбор")
		}
	}
}

func createURL(scanner *bufio.Scanner) {
	fmt.Print("\nВведите url для сокращения: ")
	scanner.Scan()
	url := strings.TrimSpace(scanner.Text())
	if url == "" {
		return
	}

	fmt.Print("Нажмите Enter для генерации короткой ссылки или введите собственное название сокращенной ссылки: ")
	scanner.Scan()
	alias := strings.TrimSpace(scanner.Text())

	body, _ := json.Marshal(map[string]string{"url": url, "alias": alias})
	resp, err := http.Post(base+"/url", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Сервер недоступен. Запустите server.exe сначала.")
		return
	}
	defer resp.Body.Close()

	var result struct {
		Alias      string `json:"alias"`
		RespStatus struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.RespStatus.Status != "OK" {
		fmt.Println("Ошибка:", result.RespStatus.Error)
		return
	}

	fmt.Printf("Успешно: %s/%s\n", base, result.Alias)
}

func deleteURL(scanner *bufio.Scanner, token string) {
	fmt.Print("\nВведите алиас для удаления: ")
	scanner.Scan()
	alias := strings.TrimSpace(scanner.Text())
	if alias == "" {
		return
	}

	req, err := http.NewRequest(http.MethodDelete, base+"/url/"+alias, nil)
	if err != nil {
		fmt.Println("Ошибка создания запроса:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Сервер недоступен. Запусти server.exe сначала.")
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		fmt.Printf("Ссылка '%s' удалена\n", alias)
	case http.StatusUnauthorized:
		fmt.Println("Ошибка: неверный токен авторизации")
	case http.StatusNotFound:
		fmt.Printf("Ошибка: алиас '%s' не найден\n", alias)
	default:
		fmt.Printf("Ошибка: сервер вернул статус %d\n", resp.StatusCode)
	}
}
