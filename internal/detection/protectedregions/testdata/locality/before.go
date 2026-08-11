package fixture

// handwritten outside generated regions
var Keep = 7

//gooo:protected:start id="fixture/handwritten"
const Protected = "keep"

//gooo:protected:end id="fixture/handwritten"

//gooo:generated:start id="fixture/activity" kind="activity"
func Activity() int {
	//gooo:slot:start id="fixture/activity/implementation"
	return 7
	//gooo:slot:end id="fixture/activity/implementation"
}

//gooo:generated:end id="fixture/activity"

//gooo:generated:start id="fixture/entity" kind="entity"
type Entity struct{}

//gooo:generated:end id="fixture/entity"
