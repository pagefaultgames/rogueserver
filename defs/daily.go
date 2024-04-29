// Copyright (C) 2024 Pagefault Games - All Rights Reserved
// https://github.com/pagefaultgames

package defs

type DailyRanking struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Score    int    `json:"score"`
	Wave     int    `json:"wave"`
}
