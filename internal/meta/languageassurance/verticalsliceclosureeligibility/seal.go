package verticalsliceclosureeligibility

import "encoding/json"

func digestJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return digestBytes(encoded)
}

func sealReport(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func sealSuite(suite Suite) Suite {
	suite.SuiteDigest = ""
	suite.SuiteDigest = digestJSON(suite)
	return suite
}
