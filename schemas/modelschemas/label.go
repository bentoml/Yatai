package modelschemas

type LabelItemSchema struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

type CreateLabelForResourceSchema struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CreateLabelsForResourceSchema []CreateLabelForResourceSchema
