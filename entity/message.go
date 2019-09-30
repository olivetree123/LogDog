package entity

type Message struct {
	Service   string
	Source    string
	Table     string
	Container string
	AddData   map[string]interface{}
	Message   interface{}
	HostName  string
}

func NewMessage(source string, table string, container string, addData map[string]interface{}, message map[string]interface{}) *Message {
	msg := Message{
		Service:   "",
		Source:    source,
		Table:     table,
		Container: container,
		AddData:   addData,
		Message:   message,
	}
	return &msg
}
