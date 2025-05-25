package main

import (
	"time"
	"net/http"
)

type Tweet struct {
	ID        string `json:"tweet_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	Username  string `json:"username"`
}

type TimelineResponse struct {
	Timeline []Tweet `json:"timeline"`
}

type App struct {
	/*
	screen         tcell.Screen
	tweets         []Tweet
	selectedIdx    int
	currentPage    int
	shortTweetsOnly bool
	longTweetsOnly  bool
	accountsList    string
	filterUsername  string
	*/
}

type Fetcher struct {
	client     *http.Client
	rateLimiter chan time.Time
	accountsList string
}
