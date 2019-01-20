package test

type MessageMock struct {
	PayloadData []byte
	TopicData   string
}

func (*MessageMock) Duplicate() bool {
	panic("implement me")
}

func (*MessageMock) Qos() byte {
	panic("implement me")
}

func (*MessageMock) Retained() bool {
	panic("implement me")
}

func (m *MessageMock) Topic() string {
	return m.TopicData
}

func (*MessageMock) MessageID() uint16 {
	panic("implement me")
}

func (m *MessageMock) Payload() []byte {
	return m.PayloadData
}
