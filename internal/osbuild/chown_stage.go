package osbuild

type ChownStageOptions struct {
	Options   []ChownStageOption `json:"options,omitempty"`
}

func (ChownStageOptions) isStageOptions() {}

type ChownStageOption struct {
	Paths []string `json:"paths"`
	Recursive bool `json:"recursive"`
	Groupname string `json:"groupname"`
	Username string `json:"username"`
}

func NewChownStage(options *ChownStageOptions) *Stage {
	return &Stage{
		Type:    "org.osbuild.chown",
		Options: options,
	}
}
