package v1alpha1

// RegisterDeployBody represents the corresponding Sleuth API endpoint:
// Deploy Registration
// Register your deploys via the Sleuth API.
// See more: https://help.sleuth.io/sleuth-api
type RegisterDeployBody struct {
	// If the value is provided and set to "true" Sleuth won't return a 400 if we see a SHA that has already been registered
	IgnoreIfDuplicate bool `json:"ignore_if_duplicate,omitempty"`
	// A comma-delimited list of tags. Defaults to tags calculated by matching paths defined in your .sleuth/TAGS file
	Tags string `json:"tags,omitempty"`
	// String defining the environment to register the deploy against. If not provided Sleuth will use the default environment of the Project
	Environment string `json:"environment,omitempty"`
	// Email address of author
	Email string `json:"email,omitempty"`
	// ISO 8601 deployment date string
	Date string `json:"date,omitempty"`
	// Located in the `Organization Settings > Details > Api Key` field
	ApiKey string `json:"api_key,omitempty"`
	// The git SHA of the deployment, located in the deploy card of the deployment
	Sha string `json:"sha,omitempty"`
	// A key/value pair which is the link name and the link itself of the form: mylink=http://my.link
	// If you need to send multiple then send a JSON body POST and this is a dictionary of values.
	Links map[string]string `json:"links,omitempty"`
}
