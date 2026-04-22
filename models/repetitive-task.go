package models

type RepetitiveTaskShort struct {
	Description  string   `json:"description" example:"Fix login bug"`
	Tags         []string `json:"tags" example:"bug,urgent"`
	Notes        string   `json:"notes" example:"User cannot log in"`
	OnceInDays   *int     `json:"onceInDays" example:"7"`
	OnceInWeeks  *int     `json:"onceInWeeks" example:"2"`
	OnceInMonths *int     `json:"onceInMonths" example:"3"`
}

