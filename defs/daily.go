package defs

type DailyRanking struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}
