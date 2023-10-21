package devcr

func dockerPodmanBuildArgs(tag string, push bool, sourcePath, file string) []string {
	args := []string{"build"}
	if tag != "" {
		args = append(args, "--tag", tag)
	}
	if push {
		args = append(args, "--push")
	}
	if file != "" {
		args = append(args, "--file", file)
	}
	args = append(args, sourcePath)

	return args
}

func dockerPodmanLoadArgs(sourcePath string) []string {
	return []string{"load", "--input", sourcePath}
}

func dockerPodmanSaveArgs(dstPath string, tags []string) []string {
	return append([]string{"save", "--output", dstPath}, tags...)
}

func dockerPodmanLoginArgs(registry, user, password string) []string {
	return []string{"login", "--username", user, "--password", password, registry}
}

func dockerPodmanPushArgs(tag string) []string {
	return []string{"push", tag}
}
