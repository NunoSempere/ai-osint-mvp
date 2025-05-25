package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func getAccounts(accountsList string) ([]string, error){
	accountsDir := "./data"
	accountsPath := filepath.Join(accountsDir, accountsList+".txt")
	accountsData, err := os.ReadFile(accountsPath)
	if err != nil {
		return []string{}, fmt.Errorf("error reading accounts file: %v", err)
	}
	accounts := strings.Split(strings.TrimSpace(string(accountsData)), "\n")

	// Filter out empty and commented accounts
	var validAccounts []string
	for _, account := range accounts {
		if account = strings.TrimSpace(account); account != "" && !strings.HasPrefix(account, "#") {
			validAccounts = append(validAccounts, account)
		}
	}
	return validAccounts, nil
}

func (a *App) loadTweets(accountsList string) ([]Tweet, error) {
	if err := godotenv.Load(".env"); err != nil {
		return []Tweet{}, fmt.Errorf("error loading .env file: %v", err)
	}

	url := os.Getenv("DATABASE_POOL_URL")
	if url == "" {
		return []Tweet{}, fmt.Errorf("DATABASE_POOL_URL environment variable not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return []Tweet{}, fmt.Errorf("in client, failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT tweet_id, tweet_text, username, created_at FROM tweets0x001 WHERE created_at >= NOW() - INTERVAL '1 day' ORDER BY created_at DESC")
	if err != nil {
		return []Tweet{}, fmt.Errorf("failed to query tweets: %v", err)
	}
	defer rows.Close()

	validAccounts, err := getAccounts(accountsList)
	if err != nil {
		return []Tweet{}, fmt.Errorf("didn't get accounts: %v", err)
	}

	var sources []Tweet
	for rows.Next() {
		var s Tweet
		var date time.Time
		err := rows.Scan(&s.ID, &s.Text, &s.Username, &date)
		s.CreatedAt = date.Format("2006-01-02 15:04:05")
		if err != nil {
			return []Tweet{}, fmt.Errorf("failed to scan row: %v", err)
		}
		if slices.Contains(validAccounts, s.Username) {
			sources = append(sources, s)
		}
	}

	return sources, nil
}
