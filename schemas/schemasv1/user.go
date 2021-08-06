package schemasv1

type UserSchema struct {
	ResourceSchema
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatar_url"`
}

type UserListSchema struct {
	BaseListSchema
	Items []*UserSchema `json:"items"`
}

type RegisterUserSchema struct {
	Name      string `json:"name" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required"`
}

type LoginUserSchema struct {
	NameOrEmail string `json:"name_or_email" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

type UpdateUserSchema struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}
