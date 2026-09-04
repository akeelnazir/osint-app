package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Comment struct {
	ent.Schema
}

func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.Text("body").NotEmpty(),
		field.JSON("mentions", []string{}).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("comments").
			Unique().
			Required(),
		edge.From("case", CaseRecord.Type).
			Ref("comments").
			Unique(),
		edge.From("evidence", Evidence.Type).
			Ref("comments").
			Unique(),
	}
}
