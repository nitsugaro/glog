package glog

type BasePaginatedResult struct {
	ResultCount int  `json:"resultCount"`
	Stop        bool `json:"stop"`
	Result      any  `json:"result"`
}
