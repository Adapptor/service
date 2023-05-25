// Adapptor helpers for writing web services
package service

type ServerType int

const (
	Production ServerType = iota
	Staging
	Development
	LiveTest
	UAT
	Local
)

func (s ServerType) String() string {
	return [...]string{"Production", "Staging", "Development", "LiveTest", "UAT", "Local"}[s]
}
