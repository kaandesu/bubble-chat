package messages

type NavigateTo struct {
	To int
}

type SetActive struct {
	Value bool
}

type UpdateUsername struct {
	Value string
}
