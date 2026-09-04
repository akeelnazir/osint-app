// Command seed populates the database with demo data for development.
package main

import (
	"context"
	"log"
	"time"

	"github.com/akeelnazir/osint-app/backend/internal/auth"
	"github.com/akeelnazir/osint-app/backend/internal/config"
	"github.com/akeelnazir/osint-app/backend/internal/ent"
	entcaserecord "github.com/akeelnazir/osint-app/backend/internal/ent/caserecord"
	entevidence "github.com/akeelnazir/osint-app/backend/internal/ent/evidence"
	entuser "github.com/akeelnazir/osint-app/backend/internal/ent/user"
	"github.com/akeelnazir/osint-app/backend/internal/geo"
	"github.com/akeelnazir/osint-app/backend/internal/migrate"

	_ "github.com/lib/pq" // register the postgres driver for database/sql

	"entgo.io/ent/dialect/sql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	drv, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	// Run migrations.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("schema create: %v", err)
	}
	if err := migrate.Apply(ctx, drv); err != nil {
		log.Printf("postgis migration warning: %v", err)
	}

	// Create admin user.
	adminHash, _ := auth.HashPassword("adminpass123")
	admin, err := client.User.Create().
		SetEmail("admin@osint.local").
		SetUsername("admin").
		SetPasswordHash(adminHash).
		SetRole(entuser.RoleAdmin).
		Save(ctx)
	if err != nil {
		log.Printf("admin user may already exist: %v", err)
		admin, _ = client.User.Query().Where(entuser.UsernameEQ("admin")).Only(ctx)
	}

	analystHash, _ := auth.HashPassword("analystpass123")
	analyst, err := client.User.Create().
		SetEmail("analyst@osint.local").
		SetUsername("analyst").
		SetPasswordHash(analystHash).
		SetRole(entuser.RoleAnalyst).
		Save(ctx)
	if err != nil {
		log.Printf("analyst user may already exist: %v", err)
		analyst, _ = client.User.Query().Where(entuser.UsernameEQ("analyst")).Only(ctx)
	}

	// Create a sample case.
	case1, err := client.CaseRecord.Create().
		SetTitle("Sample Investigation: Border Crossing Activity").
		SetDescription("Analysis of vehicle movements at a border crossing using open-source satellite imagery and social media posts.").
		SetStatus(entcaserecord.StatusOpen).
		SetVisibility(entcaserecord.VisibilityPrivate).
		SetOwnerID(admin.ID).
		Save(ctx)
	if err != nil {
		log.Fatalf("create case: %v", err)
	}

	// Add analyst as a member.
	_, _ = client.CaseMember.Create().
		SetCaseID(case1.ID).
		SetUserID(analyst.ID).
		SetRole("editor").
		Save(ctx)

	// Create sample evidence.
	lat1, lng1 := 50.4501, 30.5234 // Kyiv
	ev1, err := client.Evidence.Create().
		SetCaseID(case1.ID).
		SetType(entevidence.TypeText).
		SetTitle("Field observation report").
		SetContent("On 2024-03-15, observed convoy of 12 vehicles moving north on highway M-01. Vehicles included military transport trucks and at least 2 armored personnel carriers.").
		SetSource("Local witness report").
		SetEvidenceDate(time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)).
		SetLatitude(lat1).
		SetLongitude(lng1).
		SetDescription("Witness account of military convoy movement.").
		SetCreatorID(analyst.ID).
		Save(ctx)
	if err != nil {
		log.Printf("create evidence 1: %v", err)
	}
	if ev1 != nil {
		_ = geo.SetGeom(ctx, drv, ev1.ID, &lat1, &lng1)
	}

	lat2, lng2 := 49.8397, 24.0297 // Lviv
	ev2, err := client.Evidence.Create().
		SetCaseID(case1.ID).
		SetType(entevidence.TypeURL).
		SetTitle("Social media post with geolocation").
		SetContent("https://twitter.com/example/status/123456789").
		SetSource("Twitter/X").
		SetEvidenceDate(time.Date(2024, 3, 16, 9, 0, 0, 0, time.UTC)).
		SetLatitude(lat2).
		SetLongitude(lng2).
		SetDescription("Geolocated social media post showing military equipment.").
		SetMetadata(map[string]any{"title": "Example Tweet", "preview_image": "https://example.com/image.jpg"}).
		SetCreatorID(analyst.ID).
		Save(ctx)
	if err != nil {
		log.Printf("create evidence 2: %v", err)
	}
	if ev2 != nil {
		_ = geo.SetGeom(ctx, drv, ev2.ID, &lat2, &lng2)
	}

	// Add a comment.
	_, _ = client.Comment.Create().
		SetBody("Initial analysis suggests the convoy originated from the southwest. Cross-referencing with satellite imagery recommended. @analyst can you verify?").
		SetAuthorID(admin.ID).
		SetCaseID(case1.ID).
		SetMentions([]string{"analyst"}).
		Save(ctx)

	log.Println("Seed data created successfully!")
	log.Println("  Admin login:    admin@osint.local / adminpass123")
	log.Println("  Analyst login:  analyst@osint.local / analystpass123")
	log.Printf("  Sample case ID: %d", case1.ID)
}
