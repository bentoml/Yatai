package utils

func BoolPtr(b bool) *bool {
	return &b
}

func IntPtr(i int) *int {
	return &i
}

func UintPtr(i uint) *uint {
	return &i
}

func UintSlicePtr(i []uint) *[]uint {
	return &i
}

func StringSlicePtr(i []string) *[]string {
	return &i
}

func Int32Ptr(i int32) *int32 {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}
