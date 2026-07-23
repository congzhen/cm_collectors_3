package datatype

type PerformerAvatarStrategy string

const (
	PerformerAvatarStrategyRecommended PerformerAvatarStrategy = "recommended"
	PerformerAvatarStrategyOriginal    PerformerAvatarStrategy = "original"
	PerformerAvatarStrategyAIFix       PerformerAvatarStrategy = "aiFix"
)

type PerformerAvatarLibrarySetting struct {
	CustomBaseURL   string                  `json:"customBaseUrl"`
	DefaultStrategy PerformerAvatarStrategy `json:"defaultStrategy"`
}

type ReqParam_PerformerAvatarApply struct {
	PerformerID string `json:"performerId" binding:"required"`
	CandidateID string `json:"candidateId" binding:"required"`
	Overwrite   bool   `json:"overwrite"`
}

type ReqParam_PerformerAvatarBatch struct {
	PerformerBasesID string                  `json:"performerBasesId" binding:"required"`
	AllPerformers    bool                    `json:"allPerformers"`
	PerformerIDs     []string                `json:"performerIds"`
	Strategy         PerformerAvatarStrategy `json:"strategy" binding:"required,oneof=recommended original aiFix"`
	Overwrite        bool                    `json:"overwrite"`
}
