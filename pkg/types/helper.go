package types

func ToBool(b *bool) bool {
	return *b
}

func Bool(b bool) *bool {
	return &b
}

func ToInt(i *int) int {
	return *i
}

func Int(i int) *int {
	return &i
}

func ToString(s *string) string {
	return *s
}

func String(s string) *string {
	return &s
}
