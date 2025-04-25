package simpleim

const eventName = "simple_im_msg"

type Event struct {
	Msg      Message
	Receiver int64
}
