package changeset

import "time"

type ChangesetType string

const (
	MorphSpecType ChangesetType = "MorphSpec"
	AutopilotType ChangesetType = "Autopilot"
)

type ChangesetStatus string

const (
	Draft     ChangesetStatus = "Draft"
	Published ChangesetStatus = "Published"
	Running   ChangesetStatus = "Running"
	Completed ChangesetStatus = "Completed"
	Failed    ChangesetStatus = "Failed"
)

type Changeset struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Type         ChangesetType   `json:"type"`
	Status       ChangesetStatus `json:"status"`
	Content      string          `json:"content"` // MorphSpec content or AI prompt
	AIModel      string          `json:"ai_model,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Repositories []Repository    `gorm:"many2many:changeset_repositories;" json:"repositories"`
}

type Repository struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	PRLink   string `json:"pr_link,omitempty"`
	PRStatus string `json:"pr_status,omitempty"`
}
