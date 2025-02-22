package models

type Label struct {
	ID        int    `json:"id" gorm:"primary_key"`
	Name      string `json:"name" gorm:"column:name;not null"`
	Color     string `json:"color" gorm:"type:varchar(7);column:color;not null"`
	CreatedBy int    `json:"created_by" gorm:"column:created_by;not null"`

	User  User   `json:"user" gorm:"foreignKey:CreatedBy"`
	Tasks []Task `json:"-" gorm:"many2many:task_labels;constraint:OnDelete:CASCADE"`
}
