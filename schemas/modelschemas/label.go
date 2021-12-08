package modelschemas

type LabelItemSchema struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LabelItemsSchema []LabelItemSchema
