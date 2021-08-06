package models

type ResourceMixin struct {
	BaseModel
	Name string `json:"name"`
}

func (m *ResourceMixin) GetName() string {
	return m.Name
}
