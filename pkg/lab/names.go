package lab

type RunNames struct {
	SystemNamespace    string
	WorkspaceNamespace string
}

func NewRunNames(id, labName string) RunNames {
	return RunNames{
		SystemNamespace:    "lab-" + id + "-system",
		WorkspaceNamespace: "lab-" + id + "-workspace",
	}
}
