package terminal

// For now, this just has a printline method. However, the aim is eventually to be able to manipulate the CALFAX terminal from BASIC.
type Terminal interface {
	Printline(string)
}
