package cmd

func Pod(name string, namespace string, ContainerName ...string) pod {
	if len(ContainerName) == 0 {
		return pod{name, namespace, ""}
	} else if len(ContainerName) == 1 {
		return pod{name, namespace, ContainerName[0]}
	} else {
		panic("none or less one container")
	}
}
