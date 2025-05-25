package main

import (
        "fmt"
        "os"
        "time"
)

func main() {
        // Set the accounts list filename
        accountsList := "ai-og"

        // Get the current time and calculate the start date for the report (last 7 days)
        endDate := time.Now()
        startDate := endDate.AddDate(0, 0, -7)

        // Create an instance of App
        app := &App{}

        // Generate and save reports for the specified accounts list from the start date for 7 days
        if err := app.GenerateReports(accountsList, startDate, 7); err != nil {
                fmt.Printf("Error generating reports: %v\n", err)
                os.Exit(1)
        }

        fmt.Println("Weekly report generation completed successfully.")
}

