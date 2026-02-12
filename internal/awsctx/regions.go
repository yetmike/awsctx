package awsctx

// AWS regions as of 2025. Update periodically.
var awsRegions = []string{
	// US
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",

	// Canada
	"ca-central-1",
	"ca-west-1",

	// Europe
	"eu-central-1",
	"eu-central-2",
	"eu-west-1",
	"eu-west-2",
	"eu-west-3",
	"eu-north-1",
	"eu-south-1",
	"eu-south-2",

	// Asia Pacific
	"ap-east-1",
	"ap-south-1",
	"ap-south-2",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-southeast-3",
	"ap-southeast-4",
	"ap-southeast-5",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-northeast-3",

	// South America
	"sa-east-1",

	// Africa
	"af-south-1",

	// Middle East
	"me-south-1",
	"me-central-1",

	// Israel
	"il-central-1",

	// Mexico
	"mx-central-1",
}

func isValidRegion(name string) bool {
	for _, r := range awsRegions {
		if r == name {
			return true
		}
	}
	return false
}
