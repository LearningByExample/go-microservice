package data

import "fmt"

type Pet struct {
	Id   int
	Name string
	Race string
	Mod  string
}

func (p Pet) String() string {
	return fmt.Sprintf("{ Id: %d, Name: %q, Race: %q, Mod: %q }", p.Id, p.Name, p.Race, p.Mod)
}
