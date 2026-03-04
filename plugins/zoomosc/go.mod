module github.com/robomon1/robostream-zoomosc-controller

go 1.23.0

require (
	github.com/hypebeast/go-osc v0.0.0-20220308234300-cec5a8a1e5f5
	github.com/robomon1/robo-stream/sdk v0.0.0
)

// During development within the monorepo, point to the local sdk directory.
// Remove this replace directive and pin to a tagged release when publishing.
replace github.com/robomon1/robo-stream/sdk => ../../sdk
