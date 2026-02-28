module github.com/robomon1/robostream-zoom-controller

go 1.23.0

require github.com/robomon1/robo-stream/sdk v0.0.0

// During development within the monorepo, point to the local sdk directory.
// Remove this replace directive and pin to a tagged release when publishing.
replace github.com/robomon1/robo-stream/sdk => ../../sdk
