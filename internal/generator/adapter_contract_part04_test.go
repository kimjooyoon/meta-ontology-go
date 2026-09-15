package generator

func withFacts(input reflectiveGraphFixture, facts any) reflectiveGraphFixture {
	input.Facts = facts
	return input
}
