package main

func sealSuite(suite *Suite) {
	copy := *suite
	copy.SuiteDigest = ""
	suite.SuiteDigest = digestValue(copy)
}
