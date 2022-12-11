package kubernetes

// const dns1123Name = "([a-z0-9]+[-a-z0-9]*[a-z0-9])"
const dns1123Name = `(\w+)`

// varsubRegex is the regular expression used to validate
// the var names before substitution
const varsubRegex = "^[_[:alpha:]][_[:alpha:][:digit:]]*$"

const (
	noResourceErrMsg   string = "No resource called %s was registered in a previous step"
	patchContentErrMsg string = "Error while reading patch contents: %s"
	noAPIVersionErrMsg string = "Provided test case resource has an empty API Version"
	noKindErrMsg       string = "Provided test case resource has an empty Kind"
	noNameErrMsg       string = "Provided test case resource has an empty Name"
	notPodErrMsg       string = "Provided resource is not a Pod"
)

type podSessionKey struct{}
