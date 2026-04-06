package quest

import "time"

// QuestState 任务状态
type QuestState struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Stage       int         `json:"stage"`
	Completed   bool        `json:"completed"`
	Objectives  []Objective `json:"objectives"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
}

// Objective 任务目标
type Objective struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Target      int    `json:"target,omitempty"`
	Current     int    `json:"current,omitempty"`
}
