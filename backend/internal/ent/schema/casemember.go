package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// CaseMember is the join table between Case and User with a per-case role.
type CaseMember struct {
	ent.Schema
}

func (CaseMember) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").Values("owner", "editor", "viewer").Default("viewer"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (CaseMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("case", CaseRecord.Type).
			Ref("members").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("memberships").
			Unique().
			Required(),
	}
}

func (CaseMember) Indexes() []ent.Index {
	return []ent.Index{
		// One membership per (case, user) — enforced via edges Unique+Required.
	}
}
