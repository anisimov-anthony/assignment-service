package domain

type User struct {
	UserID   string `bson:"user_id" json:"user_id"`
	Username string `bson:"username" json:"username"`
	TeamName string `bson:"team_name" json:"team_name"`
	IsActive bool   `bson:"is_active" json:"is_active"`
}

type TeamMember struct {
	UserID   string `bson:"user_id" json:"user_id"`
	Username string `bson:"username" json:"username"`
	IsActive bool   `bson:"is_active" json:"is_active"`
}

type Team struct {
	TeamName string       `bson:"team_name" json:"team_name"`
	Members  []TeamMember `bson:"members" json:"members"`
}
