// ABOUTME: Relationship database operations using Office OS foundation
// ABOUTME: Handles CRUD operations and bidirectional relationship queries between contacts

package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/harperreed/pagen/models"
)

// orderContactIDs ensures contact_id_1 < contact_id_2 for consistent bidirectional storage.
func orderContactIDs(id1, id2 uuid.UUID) (uuid.UUID, uuid.UUID) {
	if id1.String() < id2.String() {
		return id1, id2
	}
	return id2, id1
}

func CreateRelationship(db *sql.DB, relationship *models.Relationship) error {
	relRepo := NewRelationshipsRepository(db)

	relationship.ID = uuid.New()
	now := time.Now()
	relationship.CreatedAt = now
	relationship.UpdatedAt = now

	// Ensure proper ordering for bidirectional lookup
	relationship.ContactID1, relationship.ContactID2 = orderContactIDs(relationship.ContactID1, relationship.ContactID2)

	rel := &Relationship{
		ID:       relationship.ID.String(),
		SourceID: relationship.ContactID1.String(),
		TargetID: relationship.ContactID2.String(),
		Type:     RelTypeKnows,
		Metadata: map[string]interface{}{
			"relationship_type": relationship.RelationshipType,
			"context":           relationship.Context,
		},
		CreatedAt: relationship.CreatedAt,
		UpdatedAt: relationship.UpdatedAt,
	}

	return relRepo.Create(context.Background(), rel)
}

func GetRelationship(db *sql.DB, id uuid.UUID) (*models.Relationship, error) {
	relRepo := NewRelationshipsRepository(db)

	rel, err := relRepo.Get(context.Background(), id.String())
	if errors.Is(err, ErrRelationshipNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	sourceID, _ := uuid.Parse(rel.SourceID)
	targetID, _ := uuid.Parse(rel.TargetID)
	relationshipID, _ := uuid.Parse(rel.ID)

	relationship := &models.Relationship{
		ID:         relationshipID,
		ContactID1: sourceID,
		ContactID2: targetID,
		CreatedAt:  rel.CreatedAt,
		UpdatedAt:  rel.UpdatedAt,
	}

	if relType, ok := rel.Metadata["relationship_type"].(string); ok {
		relationship.RelationshipType = relType
	}
	if ctx, ok := rel.Metadata["context"].(string); ok {
		relationship.Context = ctx
	}

	return relationship, nil
}

func FindRelationshipsBetween(db *sql.DB, contactID1, contactID2 uuid.UUID) ([]models.Relationship, error) {
	relRepo := NewRelationshipsRepository(db)

	// Order the IDs to match storage pattern
	orderedID1, orderedID2 := orderContactIDs(contactID1, contactID2)

	rels, err := relRepo.FindBetween(context.Background(), orderedID1.String(), orderedID2.String())
	if err != nil {
		return nil, err
	}

	var relationships []models.Relationship
	for _, rel := range rels {
		// Only return "knows" type relationships
		if rel.Type != RelTypeKnows {
			continue
		}

		sourceID, _ := uuid.Parse(rel.SourceID)
		targetID, _ := uuid.Parse(rel.TargetID)
		relationshipID, _ := uuid.Parse(rel.ID)

		relationship := models.Relationship{
			ID:         relationshipID,
			ContactID1: sourceID,
			ContactID2: targetID,
			CreatedAt:  rel.CreatedAt,
			UpdatedAt:  rel.UpdatedAt,
		}

		if relType, ok := rel.Metadata["relationship_type"].(string); ok {
			relationship.RelationshipType = relType
		}
		if context, ok := rel.Metadata["context"].(string); ok {
			relationship.Context = context
		}

		relationships = append(relationships, relationship)
	}

	return relationships, nil
}

func FindContactRelationships(db *sql.DB, contactID uuid.UUID, relationshipType string) ([]models.Relationship, error) {
	relRepo := NewRelationshipsRepository(db)

	// Get relationships where contact is source
	sourceRels, err := relRepo.FindBySource(context.Background(), contactID.String(), RelTypeKnows)
	if err != nil {
		return nil, err
	}

	// Get relationships where contact is target
	targetRels, err := relRepo.FindByTarget(context.Background(), contactID.String(), RelTypeKnows)
	if err != nil {
		return nil, err
	}

	var relationships []models.Relationship

	// Convert source relationships
	for _, rel := range sourceRels {
		// Apply relationship type filter if provided
		if relationshipType != "" {
			if metaRelType, ok := rel.Metadata["relationship_type"].(string); !ok || metaRelType != relationshipType {
				continue
			}
		}

		sourceID, _ := uuid.Parse(rel.SourceID)
		targetID, _ := uuid.Parse(rel.TargetID)
		relationshipID, _ := uuid.Parse(rel.ID)

		relationship := models.Relationship{
			ID:         relationshipID,
			ContactID1: sourceID,
			ContactID2: targetID,
			CreatedAt:  rel.CreatedAt,
			UpdatedAt:  rel.UpdatedAt,
		}

		if relType, ok := rel.Metadata["relationship_type"].(string); ok {
			relationship.RelationshipType = relType
		}
		if ctx, ok := rel.Metadata["context"].(string); ok {
			relationship.Context = ctx
		}

		relationships = append(relationships, relationship)
	}

	// Convert target relationships
	for _, rel := range targetRels {
		// Apply relationship type filter if provided
		if relationshipType != "" {
			if metaRelType, ok := rel.Metadata["relationship_type"].(string); !ok || metaRelType != relationshipType {
				continue
			}
		}

		sourceID, _ := uuid.Parse(rel.SourceID)
		targetID, _ := uuid.Parse(rel.TargetID)
		relationshipID, _ := uuid.Parse(rel.ID)

		relationship := models.Relationship{
			ID:         relationshipID,
			ContactID1: sourceID,
			ContactID2: targetID,
			CreatedAt:  rel.CreatedAt,
			UpdatedAt:  rel.UpdatedAt,
		}

		if relType, ok := rel.Metadata["relationship_type"].(string); ok {
			relationship.RelationshipType = relType
		}
		if ctx, ok := rel.Metadata["context"].(string); ok {
			relationship.Context = ctx
		}

		relationships = append(relationships, relationship)
	}

	return relationships, nil
}

func UpdateRelationship(db *sql.DB, id uuid.UUID, relType, ctx string) error {
	relRepo := NewRelationshipsRepository(db)

	rel, err := relRepo.Get(context.Background(), id.String())
	if err != nil {
		return err
	}

	rel.Metadata["relationship_type"] = relType
	rel.Metadata["context"] = ctx
	rel.UpdatedAt = time.Now()

	return relRepo.Update(context.Background(), rel)
}

func DeleteRelationship(db *sql.DB, id uuid.UUID) error {
	relRepo := NewRelationshipsRepository(db)
	err := relRepo.Delete(context.Background(), id.String())
	if errors.Is(err, ErrRelationshipNotFound) {
		// Return nil for backwards compatibility (old version used DELETE which succeeds even if row doesn't exist)
		return nil
	}
	return err
}

func GetAllRelationships(db *sql.DB) ([]models.Relationship, error) {
	relRepo := NewRelationshipsRepository(db)

	rels, err := relRepo.List(context.Background(), RelTypeKnows)
	if err != nil {
		return nil, err
	}

	var relationships []models.Relationship
	for _, rel := range rels {
		sourceID, _ := uuid.Parse(rel.SourceID)
		targetID, _ := uuid.Parse(rel.TargetID)
		relationshipID, _ := uuid.Parse(rel.ID)

		relationship := models.Relationship{
			ID:         relationshipID,
			ContactID1: sourceID,
			ContactID2: targetID,
			CreatedAt:  rel.CreatedAt,
			UpdatedAt:  rel.UpdatedAt,
		}

		if relType, ok := rel.Metadata["relationship_type"].(string); ok {
			relationship.RelationshipType = relType
		}
		if ctx, ok := rel.Metadata["context"].(string); ok {
			relationship.Context = ctx
		}

		relationships = append(relationships, relationship)
	}

	return relationships, nil
}
