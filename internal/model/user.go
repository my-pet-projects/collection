package model

type User struct {
	ID       string
	Username *string
}

func (u User) IsAuthenticated() bool {
	return u.ID != ""
}

func (u User) GetDisplayName() string {
	if u.Username != nil {
		return *u.Username
	}
	return u.ID
}
