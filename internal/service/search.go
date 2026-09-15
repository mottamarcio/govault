package service

import (
	"sort"
	"strings"

	"github.com/mottamarcio/govault/internal/domain"
)

// SearchFilter specifies filter criteria for in-memory record querying.
type SearchFilter struct {
	Query        string
	Type         domain.RecordType
	Tags         []string
	IncludeTrash bool
}

// FilterAndRankRecords filters and scores decrypted domain records in memory.
func FilterAndRankRecords(records []*domain.Record, filter SearchFilter) []*domain.Record {
	if len(records) == 0 {
		return []*domain.Record{}
	}

	query := strings.TrimSpace(strings.ToLower(filter.Query))
	reqTags := make([]string, 0, len(filter.Tags))
	for _, t := range filter.Tags {
		if tr := strings.TrimSpace(strings.ToLower(t)); tr != "" {
			reqTags = append(reqTags, tr)
		}
	}

	type scoredRecord struct {
		record *domain.Record
		score  int
	}

	var matched []scoredRecord

	for _, rec := range records {
		if rec == nil {
			continue
		}

		// Trash filter
		if !filter.IncludeTrash && rec.IsDeleted() {
			continue
		}
		if filter.IncludeTrash && !rec.IsDeleted() {
			// If exclusively querying trash, or if user specified IncludeTrash
		}

		// RecordType filter
		if filter.Type != "" && rec.Type != filter.Type {
			continue
		}

		// Tags filter (record must match all requested tags)
		if len(reqTags) > 0 {
			if !hasAllTags(rec.Tags, reqTags) {
				continue
			}
		}

		// Query filter & score calculation
		if query == "" {
			matched = append(matched, scoredRecord{record: rec, score: 0})
			continue
		}

		score := scoreRecord(rec, query)
		if score > 0 {
			matched = append(matched, scoredRecord{record: rec, score: score})
		}
	}

	// Sort by score descending, then by updated_at descending
	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].score != matched[j].score {
			return matched[i].score > matched[j].score
		}
		return matched[i].record.UpdatedAt.After(matched[j].record.UpdatedAt)
	})

	results := make([]*domain.Record, len(matched))
	for i, m := range matched {
		results[i] = m.record
	}
	return results
}

func hasAllTags(recordTags []string, requiredTags []string) bool {
	tagMap := make(map[string]struct{}, len(recordTags))
	for _, t := range recordTags {
		tagMap[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}

	for _, req := range requiredTags {
		if _, found := tagMap[req]; !found {
			return false
		}
	}
	return true
}

func scoreRecord(rec *domain.Record, query string) int {
	score := 0
	titleLower := strings.ToLower(rec.Title)

	if titleLower == query {
		score += 100
	} else if strings.HasPrefix(titleLower, query) {
		score += 50
	} else if strings.Contains(titleLower, query) {
		score += 30
	}

	// Tags match
	for _, tag := range rec.Tags {
		tagLower := strings.ToLower(tag)
		if tagLower == query {
			score += 25
		} else if strings.Contains(tagLower, query) {
			score += 15
		}
	}

	// Payload attributes match
	score += scorePayload(rec.Payload, query)

	return score
}

func scorePayload(payload any, query string) int {
	if payload == nil {
		return 0
	}

	score := 0

	switch p := payload.(type) {
	case domain.LoginPayload:
		score += matchString(p.Username, query, 20)
		score += matchString(p.URI, query, 15)
		score += matchString(p.Notes, query, 10)
		score += scoreFields(p.CustomFields, query)
	case *domain.LoginPayload:
		if p != nil {
			score += matchString(p.Username, query, 20)
			score += matchString(p.URI, query, 15)
			score += matchString(p.Notes, query, 10)
			score += scoreFields(p.CustomFields, query)
		}
	case domain.NotePayload:
		score += matchString(p.Content, query, 15)
		score += scoreFields(p.CustomFields, query)
	case *domain.NotePayload:
		if p != nil {
			score += matchString(p.Content, query, 15)
			score += scoreFields(p.CustomFields, query)
		}
	case domain.APIKeyPayload:
		score += matchString(p.Service, query, 20)
		score += matchString(p.Key, query, 15)
		score += matchString(p.Endpoint, query, 10)
		score += matchString(p.Notes, query, 10)
		score += scoreFields(p.CustomFields, query)
	case *domain.APIKeyPayload:
		if p != nil {
			score += matchString(p.Service, query, 20)
			score += matchString(p.Key, query, 15)
			score += matchString(p.Endpoint, query, 10)
			score += matchString(p.Notes, query, 10)
			score += scoreFields(p.CustomFields, query)
		}
	case domain.CustomPayload:
		score += matchString(p.Notes, query, 10)
		score += scoreFields(p.Fields, query)
	case *domain.CustomPayload:
		if p != nil {
			score += matchString(p.Notes, query, 10)
			score += scoreFields(p.Fields, query)
		}
	}

	return score
}

func scoreFields(fields []domain.Field, query string) int {
	score := 0
	for _, f := range fields {
		score += matchString(f.Key, query, 10)
		score += matchString(f.Value, query, 10)
	}
	return score
}

func matchString(val, query string, weight int) int {
	v := strings.ToLower(val)
	if v == "" {
		return 0
	}
	if v == query {
		return weight
	}
	if strings.Contains(v, query) {
		return weight / 2
	}
	return 0
}
