package changeset

import (
	"gorm.io/gorm"
)

// type Repository interface {
// 	GetAllChangesets() ([]Changeset, error)
// 	GetChangesetByID(id uint) (*Changeset, error)
// 	CreateChangeset(changeset *Changeset) error
// 	UpdateChangeset(changeset *Changeset) error
// 	DeleteChangeset(id uint) error
// }

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetAllChangesets() ([]Changeset, error) {
	var changesets []Changeset
	result := r.db.Preload("Repositories").Find(&changesets)
	return changesets, result.Error
}

func (r *GormRepository) GetChangesetByID(id uint) (*Changeset, error) {
	var changeset Changeset
	result := r.db.Preload("Repositories").First(&changeset, id)
	return &changeset, result.Error
}

func (r *GormRepository) CreateChangeset(changeset *Changeset) error {
	return r.db.Create(changeset).Error
}

func (r *GormRepository) UpdateChangeset(changeset *Changeset) error {
	return r.db.Save(changeset).Error
}

func (r *GormRepository) DeleteChangeset(id uint) error {
	return r.db.Delete(&Changeset{}, id).Error
}
