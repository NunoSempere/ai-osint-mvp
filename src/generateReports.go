package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type DailyReport struct {
	Date           string          `json:"date"`
	TotalTweets    int             `json:"total_tweets"`
	AccountReports []AccountReport `json:"account_reports"`
}

type WeeklyReport struct {
	StartDate      string        `json:"start_date"`
	EndDate        string        `json:"end_date"`
	DailyReports   []DailyReport `json:"daily_reports"`
	OverallSummary string        `json:"overall_summary"`
	TotalTweets    int           `json:"total_tweets"`
}

type AccountReport struct {
	Username   string  `json:"username"`
	TweetCount int     `json:"tweet_count"`
	Summary    string  `json:"summary"`
	Tweets     []Tweet `json:"tweets"`
}

// loadTweetsForDay fetches tweets for a specific day from the database
func (a *App) loadTweetsForDay(accountsList string, targetDate time.Time) ([]Tweet, error) {
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
		return []Tweet{}, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Format dates for SQL query (start and end of the target day)
	startOfDay := targetDate.Format("2006-01-02 00:00:00")
	endOfDay := targetDate.AddDate(0, 0, 1).Format("2006-01-02 00:00:00")

	rows, err := conn.Query(ctx, 
		"SELECT tweet_id, tweet_text, username, created_at FROM tweets0x001 WHERE created_at >= $1 AND created_at < $2 ORDER BY created_at DESC", 
		startOfDay, endOfDay)
	if err != nil {
		return []Tweet{}, fmt.Errorf("failed to query tweets for date %s: %v", targetDate.Format("2006-01-02"), err)
	}
	defer rows.Close()

	validAccounts, err := getAccounts(accountsList)
	if err != nil {
		return []Tweet{}, fmt.Errorf("didn't get accounts: %v", err)
	}

	var tweets []Tweet
	for rows.Next() {
		var tweet Tweet
		var date time.Time
		err := rows.Scan(&tweet.ID, &tweet.Text, &tweet.Username, &date)
		tweet.CreatedAt = date.Format("2006-01-02 15:04:05")
		if err != nil {
			return []Tweet{}, fmt.Errorf("failed to scan row: %v", err)
		}
		if slices.Contains(validAccounts, tweet.Username) {
			tweets = append(tweets, tweet)
		}
	}

	return tweets, nil
}

// summarizeTweetsForAccount creates an LLM-based summary for a specific account's tweets
func summarizeTweetsForAccount(tweets []Tweet, account string, date string) string {
	if len(tweets) == 0 {
		return fmt.Sprintf("No tweets found for @%s on %s.", account, date)
	}

	// Load OpenAI API key
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Sprintf("Error loading environment: %v", err)
	}
	
	openaiToken := os.Getenv("OPENAI_API_KEY")
	if openaiToken == "" {
		// Fallback to simple summary if no OpenAI key
		fmt.Printf("Not getting OPENAI_API_KEY")
		return fmt.Sprintf("@%s posted %d tweets on %s. OpenAI API key not configured for detailed analysis.", account, len(tweets), date)
	}

	// Combine all tweets into a single text for analysis
	var tweetTexts []string
	for _, tweet := range tweets {
		tweetTexts = append(tweetTexts, fmt.Sprintf("- %s", tweet.Text))
	}
	
	combinedText := fmt.Sprintf("Twitter activity for @%s on %s (%d tweets):\n\n%s", 
		account, date, len(tweets), strings.Join(tweetTexts, "\n"))

	fmt.Printf("Generating summary for @%s on %s (%d tweets)...\n", account, date, len(tweets))

	// Use LLM to summarize the tweets
	summary, err := Summarize(combinedText, openaiToken)
	if err != nil {
		// Fallback to simple summary if LLM fails
		return fmt.Sprintf("@%s posted %d tweets on %s. LLM analysis failed: %v", account, len(tweets), date, err)
	}

	fmt.Printf("✓ Summary generated for @%s on %s\n", account, date)
	fmt.Printf("--- Summary for @%s ---\n%s\n--- End Summary ---\n\n", account, summary)

	return fmt.Sprintf("@%s Activity Summary for %s:\n\n%s", account, date, summary)
}

// summarizeTweets creates an LLM-based summary of tweets for a given day
func summarizeTweets(tweets []Tweet, date string) string {
	if len(tweets) == 0 {
		return "No tweets found for this day."
	}

	// Load OpenAI API key
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Sprintf("Error loading environment: %v", err)
	}
	
	openaiToken := os.Getenv("OPENAI_API_KEY")
	if openaiToken == "" {
		// Fallback to simple summary if no OpenAI key
		return fmt.Sprintf("Found %d tweets on %s. OpenAI API key not configured for detailed analysis.", len(tweets), date)
	}

	// Group tweets by account
	accountTweets := make(map[string][]Tweet)
	for _, tweet := range tweets {
		accountTweets[tweet.Username] = append(accountTweets[tweet.Username], tweet)
	}

	// Create a comprehensive text for LLM analysis
	var analysisText strings.Builder
	analysisText.WriteString(fmt.Sprintf("Twitter activity analysis for %s:\n\n", date))
	
	for account, userTweets := range accountTweets {
		analysisText.WriteString(fmt.Sprintf("@%s (%d tweets):\n", account, len(userTweets)))
		for _, tweet := range userTweets {
			analysisText.WriteString(fmt.Sprintf("- %s\n", tweet.Text))
		}
		analysisText.WriteString("\n")
	}

	// Use LLM to analyze and summarize all the day's activity
	summary, err := Summarize(analysisText.String(), openaiToken)
	if err != nil {
		// Fallback to simple summary if LLM fails
		return fmt.Sprintf("Daily Activity Summary for %s:\nFound %d tweets from %d accounts. LLM analysis failed: %v", 
			date, len(tweets), len(accountTweets), err)
	}

	return fmt.Sprintf("Daily Activity Summary for %s:\n\n%s", date, summary)
}

// generateDailyReport creates a report for a specific day with separate account sections
func (a *App) generateDailyReport(accountsList string, targetDate time.Time) (DailyReport, error) {
	tweets, err := a.loadTweetsForDay(accountsList, targetDate)
	if err != nil {
		return DailyReport{}, fmt.Errorf("failed to load tweets for %s: %v", targetDate.Format("2006-01-02"), err)
	}

	fmt.Printf("Loaded %d tweets for %s\n", len(tweets), targetDate.Format("2006-01-02"))

	// Group tweets by account
	accountTweets := make(map[string][]Tweet)
	for _, tweet := range tweets {
		accountTweets[tweet.Username] = append(accountTweets[tweet.Username], tweet)
	}
	
	fmt.Printf("Found activity from %d accounts on %s\n", len(accountTweets), targetDate.Format("2006-01-02"))
	
	var accountReports []AccountReport
	for account, userTweets := range accountTweets {
		fmt.Printf("Processing account @%s (%d tweets)...\n", account, len(userTweets))
		summary := summarizeTweetsForAccount(userTweets, account, targetDate.Format("2006-01-02"))
		
		accountReport := AccountReport{
			Username:   account,
			TweetCount: len(userTweets),
			Summary:    summary,
			Tweets:     userTweets,
		}
		accountReports = append(accountReports, accountReport)
	}

	return DailyReport{
		Date:           targetDate.Format("2006-01-02"),
		TotalTweets:    len(tweets),
		AccountReports: accountReports,
	}, nil
}

// generateWeeklyReport creates a report spanning multiple days with disaggregated account data
func (a *App) generateWeeklyReport(accountsList string, startDate time.Time, days int) (WeeklyReport, error) {
	var dailyReports []DailyReport
	totalTweets := 0

	fmt.Printf("Generating %d-day report starting from %s...\n", days, startDate.Format("2006-01-02"))

	for i := 0; i < days; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		fmt.Printf("Processing day %d/%d: %s\n", i+1, days, currentDate.Format("2006-01-02"))
		
		dailyReport, err := a.generateDailyReport(accountsList, currentDate)
		if err != nil {
			fmt.Printf("Warning: Failed to generate report for %s: %v\n", currentDate.Format("2006-01-02"), err)
			continue
		}
		
		dailyReports = append(dailyReports, dailyReport)
		totalTweets += dailyReport.TotalTweets
	}

	// Generate overall summary
	overallSummary := generateOverallSummary(dailyReports)
	
	endDate := startDate.AddDate(0, 0, days-1)

	return WeeklyReport{
		StartDate:      startDate.Format("2006-01-02"),
		EndDate:        endDate.Format("2006-01-02"),
		DailyReports:   dailyReports,
		OverallSummary: overallSummary,
		TotalTweets:    totalTweets,
	}, nil
}

// generateOverallSummary creates an LLM-based summary across multiple daily reports
func generateOverallSummary(dailyReports []DailyReport) string {
	if len(dailyReports) == 0 {
		return "No daily reports available for summary."
	}

	fmt.Printf("Generating overall summary for %d days of reports...\n", len(dailyReports))

	// Load OpenAI API key
	if err := godotenv.Load(".env"); err != nil {
		return fmt.Sprintf("Error loading environment: %v", err)
	}
	
	openaiToken := os.Getenv("OPENAI_API_KEY")
	if openaiToken == "" {
		// Fallback to existing simple summary logic
		fmt.Printf("No OpenAI API key found, using simple summary...\n")
		return generateSimpleOverallSummary(dailyReports)
	}

	// Prepare comprehensive data for LLM analysis
	var analysisText strings.Builder
	analysisText.WriteString("MULTI-DAY TWITTER ACTIVITY ANALYSIS\n")
	analysisText.WriteString("=====================================\n\n")

	// Aggregate statistics
	totalTweets := 0
	allAccounts := make(map[string]int)
	activeDays := make(map[string]int)

	for _, report := range dailyReports {
		totalTweets += report.TotalTweets
		analysisText.WriteString(fmt.Sprintf("Day %s (%d total tweets):\n", report.Date, report.TotalTweets))
		
		for _, accountReport := range report.AccountReports {
			allAccounts[accountReport.Username] += accountReport.TweetCount
			activeDays[accountReport.Username]++
			
			analysisText.WriteString(fmt.Sprintf("  @%s: %d tweets\n", accountReport.Username, accountReport.TweetCount))
			
			// Include some sample tweets for context
			if len(accountReport.Tweets) > 0 {
				analysisText.WriteString("    Sample tweets:\n")
				sampleCount := min(3, len(accountReport.Tweets)) // Show up to 3 sample tweets
				for i := 0; i < sampleCount; i++ {
					tweet := accountReport.Tweets[i].Text
					if len(tweet) > 100 {
						tweet = tweet[:100] + "..."
					}
					analysisText.WriteString(fmt.Sprintf("    - %s\n", tweet))
				}
			}
		}
		analysisText.WriteString("\n")
	}

	analysisText.WriteString(fmt.Sprintf("\nPeriod: %s to %s (%d days)\n", 
		dailyReports[0].Date, 
		dailyReports[len(dailyReports)-1].Date, 
		len(dailyReports)))
	analysisText.WriteString(fmt.Sprintf("Total tweets: %d\n", totalTweets))
	analysisText.WriteString(fmt.Sprintf("Unique accounts: %d\n\n", len(allAccounts)))

	fmt.Printf("Sending %d characters to LLM for overall summary...\n", len(analysisText.String()))

	// Use LLM to create comprehensive summary
	summary, err := Summarize(analysisText.String(), openaiToken)
	if err != nil {
		// Fallback to simple summary if LLM fails
		fmt.Printf("LLM analysis failed: %v\nFalling back to simple summary...\n", err)
		return fmt.Sprintf("MULTI-DAY ACTIVITY SUMMARY\n==========================\n\nLLM analysis failed: %v\n\nFalling back to simple summary:\n\n%s", 
			err, generateSimpleOverallSummary(dailyReports))
	}

	fmt.Printf("✓ Overall summary generated successfully\n")
	fmt.Printf("--- Overall Summary ---\n%s\n--- End Overall Summary ---\n\n", summary)

	return summary
}

// generateSimpleOverallSummary creates a basic summary without LLM (fallback)
func generateSimpleOverallSummary(dailyReports []DailyReport) string {
	fmt.Printf("Generating simple overall summary (no LLM)...\n")
	
	var summary strings.Builder
	summary.WriteString("MULTI-DAY ACTIVITY SUMMARY\n")
	summary.WriteString("==========================\n\n")

	// Aggregate statistics by account
	totalTweets := 0
	allAccounts := make(map[string]int) // account -> tweet count
	activeDays := make(map[string]int)  // account -> days active

	for _, report := range dailyReports {
		totalTweets += report.TotalTweets
		
		for _, accountReport := range report.AccountReports {
			allAccounts[accountReport.Username] += accountReport.TweetCount
			activeDays[accountReport.Username]++
		}
	}

	summary.WriteString(fmt.Sprintf("Period: %s to %s (%d days)\n", 
		dailyReports[0].Date, 
		dailyReports[len(dailyReports)-1].Date, 
		len(dailyReports)))
	summary.WriteString(fmt.Sprintf("Total tweets analyzed: %d\n", totalTweets))
	summary.WriteString(fmt.Sprintf("Unique accounts: %d\n\n", len(allAccounts)))

	// Most active accounts
	summary.WriteString("Account Activity Breakdown:\n")
	type accountActivity struct {
		name   string
		tweets int
		days   int
	}
	
	var activities []accountActivity
	for account, tweetCount := range allAccounts {
		activities = append(activities, accountActivity{
			name:   account,
			tweets: tweetCount,
			days:   activeDays[account],
		})
	}
	
	// Simple sorting by tweet count (descending)
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[j].tweets > activities[i].tweets {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}
	
	for i, activity := range activities {
		avgTweetsPerDay := float64(activity.tweets) / float64(activity.days)
		summary.WriteString(fmt.Sprintf("  %d. @%s: %d tweets across %d days (avg %.1f tweets/day)\n", 
			i+1, activity.name, activity.tweets, activity.days, avgTweetsPerDay))
	}

	// Daily activity pattern
	summary.WriteString("\nDaily Activity Pattern:\n")
	for _, report := range dailyReports {
		summary.WriteString(fmt.Sprintf("  %s: %d total tweets\n", 
			report.Date, report.TotalTweets))
		for _, accountReport := range report.AccountReports {
			summary.WriteString(fmt.Sprintf("    - @%s: %d tweets\n", 
				accountReport.Username, accountReport.TweetCount))
		}
	}

	summary.WriteString("\nKey Observations:\n")
	if totalTweets == 0 {
		summary.WriteString("  - No tweet activity detected in the analyzed period\n")
	} else {
		avgTweetsPerDay := float64(totalTweets) / float64(len(dailyReports))
		summary.WriteString(fmt.Sprintf("  - Average tweets per day: %.1f\n", avgTweetsPerDay))
		
		if len(activities) > 0 {
			topAccount := activities[0]
			summary.WriteString(fmt.Sprintf("  - Most active account: @%s with %d tweets\n", 
				topAccount.name, topAccount.tweets))
		}
		
		// Find most active day
		maxDayTweets := 0
		maxDay := ""
		for _, report := range dailyReports {
			if report.TotalTweets > maxDayTweets {
				maxDayTweets = report.TotalTweets
				maxDay = report.Date
			}
		}
		if maxDay != "" {
			summary.WriteString(fmt.Sprintf("  - Most active day: %s with %d tweets\n", maxDay, maxDayTweets))
		}
	}

	return summary.String()
}

// Helper function for min (Go 1.21+)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// saveReportToFile saves a report to a JSON file
func saveReportToFile(report interface{}, filename string) error {
	reportsDir := "./data/reports"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %v", err)
	}

	filePath := filepath.Join(reportsDir, filename)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %v", err)
	}

	fmt.Printf("Report saved to: %s\n", filePath)
	return nil
}

// GenerateReports is the main function to generate reports
func (a *App) GenerateReports(accountsList string, startDate time.Time, days int) error {
	fmt.Printf("Starting report generation for %s accounts...\n", accountsList)
	
	// Generate the multi-day report
	weeklyReport, err := a.generateWeeklyReport(accountsList, startDate, days)
	if err != nil {
		return fmt.Errorf("failed to generate weekly report: %v", err)
	}

	// Save the full report
	reportFilename := fmt.Sprintf("report_%s_%s_to_%s.json", 
		accountsList, 
		weeklyReport.StartDate, 
		weeklyReport.EndDate)
	
	if err := saveReportToFile(weeklyReport, reportFilename); err != nil {
		return fmt.Errorf("failed to save report: %v", err)
	}

	// Print summary to console
	fmt.Println("\n" + weeklyReport.OverallSummary)
	
	// Save a text summary as well
	summaryFilename := fmt.Sprintf("summary_%s_%s_to_%s.txt", 
		accountsList, 
		weeklyReport.StartDate, 
		weeklyReport.EndDate)
	
	summaryPath := filepath.Join("./reports", summaryFilename)
	if err := os.WriteFile(summaryPath, []byte(weeklyReport.OverallSummary), 0644); err != nil {
		fmt.Printf("Warning: Failed to save text summary: %v\n", err)
	} else {
		fmt.Printf("Summary saved to: %s\n", summaryPath)
	}

	return nil
}
