package types

type Version struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	CommitID  string `json:"commitID"`
	BuildTime string `json:"buildTime"`
}
