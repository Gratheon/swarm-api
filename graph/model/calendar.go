package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const calendarRangeCapYears = 1

type Calendar struct {
	Db     *sqlx.DB
	UserID string
}

type CalendarInput struct {
	From        string                   `json:"from"`
	To          string                   `json:"to"`
	ApiaryID    *string                  `json:"apiaryId"`
	HiveID      *string                  `json:"hiveId"`
	SourceTypes []CalendarItemSourceType `json:"sourceTypes"`
}

type CalendarPayload struct {
	Range             *CalendarRange               `json:"range"`
	Items             []*CalendarItem              `json:"items"`
	InspectionRecency []*CalendarInspectionRecency `json:"inspectionRecency"`
}

type CalendarRange struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Capped bool   `json:"capped"`
}

type CalendarItem struct {
	ID                 string                  `json:"id"`
	Kind               CalendarItemKind        `json:"kind"`
	SourceType         CalendarItemSourceType  `json:"sourceType"`
	Date               string                  `json:"date"`
	Label              *CalendarItemLabel      `json:"label"`
	Details            *CalendarItemLabel      `json:"details"`
	Hive               *Hive                   `json:"hive"`
	Apiary             *Apiary                 `json:"apiary"`
	Source             *CalendarSourceContext  `json:"source"`
	TemplateKey        *string                 `json:"templateKey"`
	ReminderStateID    *string                 `json:"reminderStateId"`
	ReminderStatus     *CalendarReminderStatus `json:"reminderStatus"`
	LegalDisclaimerKey *string                 `json:"legalDisclaimerKey"`
}

type CalendarItemLabel struct {
	TranslationKey string  `json:"translationKey"`
	Fallback       string  `json:"fallback"`
	Args           *string `json:"args"`
}

type CalendarSourceContext struct {
	SourceType  CalendarItemSourceType `json:"sourceType"`
	SourceID    *string                `json:"sourceId"`
	HiveID      *string                `json:"hiveId"`
	ApiaryID    *string                `json:"apiaryId"`
	FamilyID    *string                `json:"familyId"`
	TemplateKey *string                `json:"templateKey"`
}

type CalendarInspectionRecency struct {
	Hive                  *Hive       `json:"hive"`
	LatestInspection      *Inspection `json:"latestInspection"`
	LatestAt              *string     `json:"latestAt"`
	IsInsideSelectedRange bool        `json:"isInsideSelectedRange"`
}

type CalendarItemKind string

const (
	CalendarItemKindHistoricalRecord  CalendarItemKind = "HISTORICAL_RECORD"
	CalendarItemKindGeneratedReminder CalendarItemKind = "GENERATED_REMINDER"
)

func (e CalendarItemKind) IsValid() bool {
	switch e {
	case CalendarItemKindHistoricalRecord, CalendarItemKindGeneratedReminder:
		return true
	}
	return false
}

func (e CalendarItemKind) String() string { return string(e) }

type CalendarItemSourceType string

const (
	CalendarItemSourceTypeInspection        CalendarItemSourceType = "INSPECTION"
	CalendarItemSourceTypeHiveLog           CalendarItemSourceType = "HIVE_LOG"
	CalendarItemSourceTypeTreatmentReminder CalendarItemSourceType = "TREATMENT_REMINDER"
	CalendarItemSourceTypeQueenMilestone    CalendarItemSourceType = "QUEEN_MILESTONE"
)

func (e CalendarItemSourceType) IsValid() bool {
	switch e {
	case CalendarItemSourceTypeInspection, CalendarItemSourceTypeHiveLog, CalendarItemSourceTypeTreatmentReminder, CalendarItemSourceTypeQueenMilestone:
		return true
	}
	return false
}

func (e CalendarItemSourceType) String() string { return string(e) }

type CalendarReminderStatus string

const (
	CalendarReminderStatusScheduled CalendarReminderStatus = "SCHEDULED"
	CalendarReminderStatusDone      CalendarReminderStatus = "DONE"
	CalendarReminderStatusDismissed CalendarReminderStatus = "DISMISSED"
	CalendarReminderStatusSnoozed   CalendarReminderStatus = "SNOOZED"
)

func (e CalendarReminderStatus) IsValid() bool {
	switch e {
	case CalendarReminderStatusScheduled, CalendarReminderStatusDone, CalendarReminderStatusDismissed, CalendarReminderStatusSnoozed:
		return true
	}
	return false
}

func (e CalendarReminderStatus) String() string { return string(e) }

type calendarHiveContextRow struct {
	ID               string  `db:"id"`
	UserID           string  `db:"user_id"`
	ApiaryID         int     `db:"apiary_id"`
	BoxSystemID      *int    `db:"box_system_id"`
	HiveType         string  `db:"hive_type"`
	Active           *bool   `db:"active"`
	HiveNumber       *int    `db:"hive_number"`
	Notes            *string `db:"notes"`
	Color            *string `db:"color"`
	Status           *string `db:"status"`
	Added            *string `db:"added"`
	CollapseDate     *string `db:"collapse_date"`
	CollapseCause    *string `db:"collapse_cause"`
	ParentHiveID     *int    `db:"parent_hive_id"`
	SplitDate        *string `db:"split_date"`
	MergedIntoHiveID *int    `db:"merged_into_hive_id"`
	MergeDate        *string `db:"merge_date"`
	MergeType        *string `db:"merge_type"`
	ApiaryName       *string `db:"apiary_name"`
	ApiaryType       string  `db:"apiary_type"`
	ApiaryLocation   *string `db:"apiary_location"`
	ApiaryActive     *bool   `db:"apiary_active"`
	ApiaryLat        *string `db:"apiary_lat"`
	ApiaryLng        *string `db:"apiary_lng"`
}

type calendarInspectionRow struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
	HiveID string `db:"hive_id"`
	Data   string `db:"data"`
	Added  string `db:"added"`
}

type calendarHiveLogRow struct {
	ID        string  `db:"id"`
	UserID    string  `db:"user_id"`
	HiveID    string  `db:"hive_id"`
	Action    string  `db:"action"`
	Title     string  `db:"title"`
	Details   *string `db:"details"`
	Source    *string `db:"source"`
	DedupeKey *string `db:"dedupe_key"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
}

func (r *Calendar) Payload(input CalendarInput) (*CalendarPayload, error) {
	from, to, capped, err := clampCalendarRange(input.From, input.To, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	sourceSet, err := calendarSourceSet(input.SourceTypes)
	if err != nil {
		return nil, err
	}

	hiveRows, err := r.listCalendarHives(input.ApiaryID, input.HiveID)
	if err != nil {
		return nil, err
	}

	hives := make(map[string]*Hive, len(hiveRows))
	apiaries := make(map[string]*Apiary, len(hiveRows))
	for _, row := range hiveRows {
		hive := hiveFromCalendarRow(row)
		apiary := apiaryFromCalendarRow(row)
		hives[hive.ID] = hive
		apiaries[strconv.Itoa(apiary.ID)] = apiary
	}

	payload := &CalendarPayload{
		Range: &CalendarRange{
			From:   formatCalendarTime(from),
			To:     formatCalendarTime(to),
			Capped: capped,
		},
		Items:             []*CalendarItem{},
		InspectionRecency: []*CalendarInspectionRecency{},
	}

	if len(hives) == 0 {
		return payload, nil
	}

	if sourceSet[CalendarItemSourceTypeInspection] {
		inspectionItems, err := r.listCalendarInspectionItems(from, to, hives, apiaries)
		if err != nil {
			return nil, err
		}
		payload.Items = append(payload.Items, inspectionItems...)
	}

	if sourceSet[CalendarItemSourceTypeHiveLog] {
		hiveLogItems, err := r.listCalendarHiveLogItems(from, to, hives, apiaries)
		if err != nil {
			return nil, err
		}
		payload.Items = append(payload.Items, hiveLogItems...)
	}

	recency, err := r.listInspectionRecency(from, to, hives)
	if err != nil {
		return nil, err
	}
	payload.InspectionRecency = recency

	return payload, nil
}

func clampCalendarRange(fromRaw string, toRaw string, now time.Time) (time.Time, time.Time, bool, error) {
	from, err := parseCalendarTime(fromRaw)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid calendar from date: %w", err)
	}
	to, err := parseCalendarTime(toRaw)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid calendar to date: %w", err)
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, false, errors.New("calendar to date must be after from date")
	}

	minFrom := now.AddDate(-calendarRangeCapYears, 0, 0)
	maxTo := now.AddDate(calendarRangeCapYears, 0, 0)
	capped := false
	if from.Before(minFrom) {
		from = minFrom
		capped = true
	}
	if to.After(maxTo) {
		to = maxTo
		capped = true
	}
	if to.Before(from) {
		to = from
		capped = true
	}

	return from, to, capped, nil
}

func parseCalendarTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("empty date")
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func formatCalendarTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func calendarSourceSet(sourceTypes []CalendarItemSourceType) (map[CalendarItemSourceType]bool, error) {
	if len(sourceTypes) == 0 {
		return map[CalendarItemSourceType]bool{
			CalendarItemSourceTypeInspection:        true,
			CalendarItemSourceTypeHiveLog:           true,
			CalendarItemSourceTypeTreatmentReminder: true,
			CalendarItemSourceTypeQueenMilestone:    true,
		}, nil
	}

	set := map[CalendarItemSourceType]bool{}
	for _, sourceType := range sourceTypes {
		if !sourceType.IsValid() {
			return nil, fmt.Errorf("unsupported calendar source type %q", sourceType)
		}
		set[sourceType] = true
	}
	return set, nil
}

func (r *Calendar) listCalendarHives(apiaryID *string, hiveID *string) ([]*calendarHiveContextRow, error) {
	conditions := []string{
		"h.user_id=?",
		"h.active=1",
		"a.user_id=?",
		"a.active=1",
	}
	args := []interface{}{r.UserID, r.UserID}

	if apiaryID != nil && *apiaryID != "" {
		conditions = append(conditions, "h.apiary_id=?")
		args = append(args, *apiaryID)
	}
	if hiveID != nil && *hiveID != "" {
		conditions = append(conditions, "h.id=?")
		args = append(args, *hiveID)
	}

	query := fmt.Sprintf(`
		SELECT h.id, h.user_id, h.apiary_id, h.box_system_id, h.hive_type, h.active, h.hive_number, h.notes, h.color, h.status, h.added,
		       h.collapse_date, h.collapse_cause, h.parent_hive_id, h.split_date, h.merged_into_hive_id, h.merge_date, h.merge_type,
		       a.name AS apiary_name, a.type AS apiary_type, NULL AS apiary_location, a.active AS apiary_active, a.lat AS apiary_lat, a.lng AS apiary_lng
		FROM hives h
		INNER JOIN apiaries a ON a.id = h.apiary_id
		WHERE %s`, strings.Join(conditions, " AND "))

	rows := []*calendarHiveContextRow{}
	err := r.Db.Select(&rows, query, args...)
	return rows, err
}

func (r *Calendar) listCalendarInspectionItems(from time.Time, to time.Time, hives map[string]*Hive, apiaries map[string]*Apiary) ([]*CalendarItem, error) {
	rows := []*calendarInspectionRow{}
	err := r.Db.Select(&rows,
		`SELECT i.id, i.user_id, i.hive_id, i.data, i.added
		 FROM inspections i
		 INNER JOIN hives h ON h.id = i.hive_id AND h.user_id = i.user_id AND h.active=1
		 INNER JOIN apiaries a ON a.id = h.apiary_id AND a.user_id = i.user_id AND a.active=1
		 WHERE i.user_id=? AND i.added >= ? AND i.added <= ?
		 ORDER BY i.added ASC, i.id ASC`,
		r.UserID,
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	items := []*CalendarItem{}
	for _, row := range rows {
		hive := hives[row.HiveID]
		if hive == nil {
			continue
		}
		apiary := apiaries[strconv.Itoa(hive.ApiaryID)]
		hiveID := row.HiveID
		apiaryID := strconv.Itoa(hive.ApiaryID)
		sourceID := row.ID
		items = append(items, &CalendarItem{
			ID:         "inspection:" + row.ID,
			Kind:       CalendarItemKindHistoricalRecord,
			SourceType: CalendarItemSourceTypeInspection,
			Date:       row.Added,
			Label: &CalendarItemLabel{
				TranslationKey: "calendar.item.inspection",
				Fallback:       "Inspection",
				Args:           calendarLabelArgs(hive, apiary),
			},
			Details: &CalendarItemLabel{
				TranslationKey: "calendar.item.inspection.details",
				Fallback:       "Hive inspection record",
				Args:           calendarLabelArgs(hive, apiary),
			},
			Hive:   hive,
			Apiary: apiary,
			Source: &CalendarSourceContext{
				SourceType: CalendarItemSourceTypeInspection,
				SourceID:   &sourceID,
				HiveID:     &hiveID,
				ApiaryID:   &apiaryID,
			},
		})
	}
	return items, nil
}

func (r *Calendar) listCalendarHiveLogItems(from time.Time, to time.Time, hives map[string]*Hive, apiaries map[string]*Apiary) ([]*CalendarItem, error) {
	rows := []*calendarHiveLogRow{}
	err := r.Db.Select(&rows,
		`SELECT hl.id, hl.user_id, hl.hive_id, hl.action, hl.title, hl.details, hl.source, hl.dedupe_key, hl.created_at, hl.updated_at
		 FROM hive_logs hl
		 INNER JOIN hives h ON h.id = hl.hive_id AND h.user_id = hl.user_id AND h.active=1
		 INNER JOIN apiaries a ON a.id = h.apiary_id AND a.user_id = hl.user_id AND a.active=1
		 WHERE hl.user_id=? AND hl.active=1 AND hl.created_at >= ? AND hl.created_at <= ?
		 ORDER BY hl.created_at ASC, hl.id ASC`,
		r.UserID,
		from,
		to,
	)
	if err != nil {
		return nil, err
	}

	items := []*CalendarItem{}
	for _, row := range rows {
		hive := hives[row.HiveID]
		if hive == nil {
			continue
		}
		apiary := apiaries[strconv.Itoa(hive.ApiaryID)]
		hiveID := row.HiveID
		apiaryID := strconv.Itoa(hive.ApiaryID)
		sourceID := row.ID
		items = append(items, &CalendarItem{
			ID:         "hive-log:" + row.ID,
			Kind:       CalendarItemKindHistoricalRecord,
			SourceType: CalendarItemSourceTypeHiveLog,
			Date:       row.CreatedAt,
			Label: &CalendarItemLabel{
				TranslationKey: "calendar.item.hive_log",
				Fallback:       row.Title,
				Args:           calendarHiveLogLabelArgs(row, hive, apiary),
			},
			Details: hiveLogDetailsLabel(row, hive, apiary),
			Hive:    hive,
			Apiary:  apiary,
			Source: &CalendarSourceContext{
				SourceType: CalendarItemSourceTypeHiveLog,
				SourceID:   &sourceID,
				HiveID:     &hiveID,
				ApiaryID:   &apiaryID,
			},
		})
	}
	return items, nil
}

func (r *Calendar) listInspectionRecency(from time.Time, to time.Time, hives map[string]*Hive) ([]*CalendarInspectionRecency, error) {
	rows := []*calendarInspectionRow{}
	err := r.Db.Select(&rows,
		`SELECT i.id, i.user_id, i.hive_id, i.data, i.added
		 FROM inspections i
		 INNER JOIN (
		     SELECT hive_id, MAX(CONCAT(DATE_FORMAT(added, '%Y%m%d%H%i%s'), LPAD(id, 10, '0'))) AS latest_key
		     FROM inspections
		     WHERE user_id=?
		     GROUP BY hive_id
		 ) latest ON latest.hive_id = i.hive_id
		          AND latest.latest_key = CONCAT(DATE_FORMAT(i.added, '%Y%m%d%H%i%s'), LPAD(i.id, 10, '0'))
		 WHERE i.user_id=?
		 ORDER BY i.added DESC, i.id DESC`,
		r.UserID,
		r.UserID,
	)
	if err != nil {
		return nil, err
	}

	result := []*CalendarInspectionRecency{}
	for _, row := range rows {
		hive := hives[row.HiveID]
		if hive == nil {
			continue
		}
		latestAt := row.Added
		latestTime, parseErr := parseCalendarTime(row.Added)
		inside := false
		if parseErr == nil {
			inside = !latestTime.Before(from) && !latestTime.After(to)
		}
		result = append(result, &CalendarInspectionRecency{
			Hive: hive,
			LatestInspection: &Inspection{
				ID:     row.ID,
				UserID: row.UserID,
				HiveID: row.HiveID,
				Data:   row.Data,
				Added:  row.Added,
			},
			LatestAt:              &latestAt,
			IsInsideSelectedRange: inside,
		})
	}
	return result, nil
}

func hiveFromCalendarRow(row *calendarHiveContextRow) *Hive {
	return &Hive{
		ID:               row.ID,
		UserID:           row.UserID,
		ApiaryID:         row.ApiaryID,
		BoxSystemID:      row.BoxSystemID,
		HiveType:         row.HiveType,
		Active:           row.Active,
		HiveNumber:       row.HiveNumber,
		Notes:            row.Notes,
		Color:            row.Color,
		Status:           row.Status,
		Added:            row.Added,
		CollapseDate:     row.CollapseDate,
		CollapseCause:    row.CollapseCause,
		ParentHiveID:     row.ParentHiveID,
		SplitDate:        row.SplitDate,
		MergedIntoHiveID: row.MergedIntoHiveID,
		MergeDate:        row.MergeDate,
		MergeType:        row.MergeType,
	}
}

func apiaryFromCalendarRow(row *calendarHiveContextRow) *Apiary {
	apiary := &Apiary{
		ID:       row.ApiaryID,
		UserID:   row.UserID,
		Name:     row.ApiaryName,
		Type:     ApiaryType(row.ApiaryType),
		Location: row.ApiaryLocation,
		Active:   row.ApiaryActive,
		Lat:      row.ApiaryLat,
		Lng:      row.ApiaryLng,
	}
	ensureApiaryType(apiary)
	return apiary
}

func calendarLabelArgs(hive *Hive, apiary *Apiary) *string {
	parts := []string{}
	if hive != nil && hive.HiveNumber != nil {
		parts = append(parts, fmt.Sprintf("\"hiveNumber\":%d", *hive.HiveNumber))
	}
	if hive != nil {
		parts = append(parts, fmt.Sprintf("\"hiveId\":\"%s\"", hive.ID))
	}
	if apiary != nil && apiary.Name != nil {
		parts = append(parts, fmt.Sprintf("\"apiaryName\":%q", *apiary.Name))
	}
	if apiary != nil {
		parts = append(parts, fmt.Sprintf("\"apiaryId\":\"%d\"", apiary.ID))
	}
	json := "{" + strings.Join(parts, ",") + "}"
	return &json
}

func calendarHiveLogLabelArgs(row *calendarHiveLogRow, hive *Hive, apiary *Apiary) *string {
	base := calendarLabelArgs(hive, apiary)
	trimmed := strings.TrimSuffix(strings.TrimPrefix(*base, "{"), "}")
	parts := []string{}
	if trimmed != "" {
		parts = append(parts, trimmed)
	}
	parts = append(parts, fmt.Sprintf("\"title\":%q", row.Title))
	parts = append(parts, fmt.Sprintf("\"action\":%q", row.Action))
	json := "{" + strings.Join(parts, ",") + "}"
	return &json
}

func hiveLogDetailsLabel(row *calendarHiveLogRow, hive *Hive, apiary *Apiary) *CalendarItemLabel {
	if row.Details == nil || *row.Details == "" {
		return nil
	}
	return &CalendarItemLabel{
		TranslationKey: "calendar.item.hive_log.details",
		Fallback:       *row.Details,
		Args:           calendarHiveLogLabelArgs(row, hive, apiary),
	}
}
