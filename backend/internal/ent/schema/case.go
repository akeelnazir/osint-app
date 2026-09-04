package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CaseRecord models an investigation case. The Go type is named CaseRecord
// (not Case) to avoid the lowercase "case" keyword conflict in generated code;
// the SQL table is still named "cases" via the entsql table annotation.
type CaseRecord struct {
	ent.Schema
}

func (CaseRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cases"},
	}
}

func (CaseRecord) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty(),
		field.Text("description").Optional(),
		field.Enum("status").Values("open", "closed", "archived").Default("open"),
		field.Enum("visibility").Values("private", "public").Default("private"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (CaseRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("owned_cases").
			Unique().
			Required(),
		edge.To("members", CaseMember.Type),
		edge.To("evidence", Evidence.Type),
		edge.To("comments", Comment.Type),
		edge.To("tags", Tag.Type),
		edge.To("activity", ActivityLog.Type),
	}
}

func (CaseRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("visibility"),
		index.Fields("updated_at"),
	}
}
