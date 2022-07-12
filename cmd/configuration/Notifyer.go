package configuration

type Notifyer interface {
	Callback(*ConfigHandler)
}
