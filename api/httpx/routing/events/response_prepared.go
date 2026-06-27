package events

// ResponsePrepared is dispatched after a response has been prepared and is
// ready to be returned to the client.
type ResponsePrepared struct {
	Request  any // foundation.Request
	Response any // foundation.Response
}
