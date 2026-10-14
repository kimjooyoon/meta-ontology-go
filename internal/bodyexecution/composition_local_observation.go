package bodyexecution

const localObservationFile = "gooo-native-local-observation.json"

type CompositionLocalWrite struct {
	SiteIndex    int  `json:"site_index"`
	Frame        int  `json:"frame"`
	ReadStart    int  `json:"read_start"`
	ReadEnd      int  `json:"read_end"`
	UseStart     int  `json:"use_start"`
	UseEnd       int  `json:"use_end"`
	ReadsDropped int  `json:"reads_dropped"`
	UsesDropped  int  `json:"uses_dropped"`
	Completed    bool `json:"completed"`
}

type CompositionLocalUse struct {
	SiteIndex  int    `json:"site_index"`
	Frame      int    `json:"frame"`
	Assignment int    `json:"assignment"`
	Reason     string `json:"reason,omitempty"`
}

type CompositionLocalElementWindow struct {
	Construction int `json:"construction"`
	Element      int `json:"element"`
	UseStart     int `json:"use_start"`
	UseEnd       int `json:"use_end"`
	UsesDropped  int `json:"uses_dropped"`
}

type CompositionLocalSummary struct {
	Schema         string `json:"schema"`
	Status         string `json:"status"`
	SHA256         string `json:"sha256"`
	RecordedWrites int    `json:"recorded_writes"`
	RecordedUses   int    `json:"recorded_uses"`
	ElementWindows int    `json:"element_windows"`
	Interrupted    int    `json:"interrupted_writes"`
	Unresolved     int    `json:"unresolved_uses"`
	Frames         int    `json:"frames"`
	FramesDropped  int    `json:"frames_dropped"`
	WritesDropped  int    `json:"writes_dropped"`
	UsesDropped    int    `json:"uses_dropped"`
	ReadsDropped   int    `json:"reads_dropped"`
	Scope          string `json:"scope"`
}
