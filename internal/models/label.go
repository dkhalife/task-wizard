package models

type Label struct {
	ID        int    `json:"id" gorm:"primary_key"`
	Name      string `json:"name" gorm:"column:name"`
	Color     string `json:"color" gorm:"type:varchar(7);column:color"`
	CreatedBy int    `json:"created_by" gorm:"column:created_by"`

	User  User   `json:"user" gorm:"foreignKey:CreatedBy"`
	Tasks []Task `json:"-" gorm:"many2many:task_labels;constraint:OnDelete:CASCADE"`
}
