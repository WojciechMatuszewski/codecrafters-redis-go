package redis

// type MessageType string

// const (
// 	Echo        MessageType = "echo"
// 	Ping        MessageType = "ping"
// 	Set         MessageType = "set"
// 	Get         MessageType = "get"
// 	Cfg         MessageType = "config"
// 	Info        MessageType = "info"
// 	Pong        MessageType = "pong"
// 	ReplicaConf MessageType = "replconf"
// 	PSync       MessageType = "psync"
// )

// type Message struct {
// 	Type   MessageType
// 	Values []string
// }

// func ParseMessage(buf []byte) Message {
// 	resp := NewResp(bytes.NewReader(buf))
// 	value, err := resp.Read()
// 	if err != nil {
// 		panic(err)
// 	}

// 	return fromValue(value)
// }

// func fromValue(value Value) Message {
// 	switch value.Type {
// 	case Array:
// 		mType := strings.ToLower(value.Array[0].Bulk)
// 		var values []string
// 		for i := 1; i < len(value.Array); i++ {
// 			values = append(values, value.Array[i].Bulk)
// 		}

// 		return Message{Type: MessageType(mType), Values: values}

// 	case SimpleString:
// 		mType := strings.ToLower(value.SimpleString)
// 		return Message{Type: MessageType(mType), Values: []string{}}

// 	case Bulk:
// 		mType := strings.ToLower(value.Bulk)
// 		return Message{Type: MessageType(mType), Values: []string{}}

// 	default:
// 		panic("Unknown value")
// 	}
// }
