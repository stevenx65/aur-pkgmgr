package main

type Package struct {
	Name             string
	Version          string
	Description      string
	Repository       string
	InstalledVersion string
}

type Tab int

const (
	TabSearch Tab = iota
	TabInstalled
	TabUpdates
)

type InputMode int

const (
	ModeNavigate InputMode = iota
	ModeSearch
	ModeFilter
)

type MsgKind int

const (
	MsgSearchResult MsgKind = iota
	MsgInstalledList
	MsgUpdatesList
	MsgPackageInfo
	MsgCmdResult
	MsgStatus
	MsgError
)

type Msg struct {
	Kind    MsgKind
	Payload interface{}
}
