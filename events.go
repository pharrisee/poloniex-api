package poloniex

import "github.com/chuckpreslar/emission"

//On adds a listener to a specific event
func (p *Poloniex) On(event interface{}, listener interface{}) *emission.Emitter {
	return p.emitter.On(event, listener)
}

//Emit emits an event
func (p *Poloniex) Emit(event interface{}, arguments ...interface{}) *emission.Emitter {
	return p.emitter.Emit(event, arguments...)
}

//Off removes a listener for an event
func (p *Poloniex) Off(event interface{}, listener interface{}) *emission.Emitter {
	return p.emitter.Off(event, listener)
}
