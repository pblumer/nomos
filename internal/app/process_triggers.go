package app

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
)

// Trigger type constants kept in one place so the validation and DTO
// translation cannot drift.
const (
	TriggerTypeNone        = "none"
	TriggerTypeTimer       = "timer"
	TriggerTypeMessage     = "message"
	TriggerTypeSignal      = "signal"
	TriggerTypeConditional = "conditional"
)

// StartEventInfo describes a BPMN start event extracted from process XML,
// including the canonical trigger type derived from its child event
// definitions. See ADR-0018.
type StartEventInfo struct {
	ID   string
	Name string
	Type string
}

// ExtractBPMNStartEvents walks a BPMN XML document and returns every
// start event together with its detected trigger type. Plain start events
// without an event definition map to TriggerTypeNone.
func ExtractBPMNStartEvents(xmlText string) ([]StartEventInfo, error) {
	dec := xml.NewDecoder(bytes.NewBufferString(xmlText))
	events := []StartEventInfo{}
	var current *StartEventInfo
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "startEvent" {
				ev := StartEventInfo{Type: TriggerTypeNone}
				for _, a := range t.Attr {
					if a.Name.Local == "id" {
						ev.ID = a.Value
					}
					if a.Name.Local == "name" {
						ev.Name = a.Value
					}
				}
				current = &ev
				continue
			}
			if current == nil {
				continue
			}
			switch t.Name.Local {
			case "timerEventDefinition":
				current.Type = TriggerTypeTimer
			case "messageEventDefinition":
				current.Type = TriggerTypeMessage
			case "signalEventDefinition":
				current.Type = TriggerTypeSignal
			case "conditionalEventDefinition":
				current.Type = TriggerTypeConditional
			}
		case xml.EndElement:
			if t.Name.Local == "startEvent" && current != nil {
				if current.ID != "" {
					events = append(events, *current)
				}
				current = nil
			}
		}
	}
	return events, nil
}

// GetProcessTriggers returns the merged trigger view (BPMN start events ∪
// YAML configuration) for the given process.
func GetProcessTriggers(path, id string) ([]ProcessTriggerDTO, error) {
	dto, err := GetProcess(path, id)
	if err != nil {
		return nil, err
	}
	return dto.Triggers, nil
}

// UpdateProcessTriggers replaces the trigger configuration on a process.
// Triggers referencing unknown BPMN elements are rejected outright; type
// mismatches and config issues surface as validation findings instead.
func UpdateProcessTriggers(path, id string, req UpdateProcessTriggersRequest) (ProcessDTO, error) {
	node, err := findProcessNode(path, id)
	if err != nil {
		return ProcessDTO{}, err
	}
	bpmnPath := safeBPMNPath(filepath.Dir(node.Path), node.Meta.BPMN.File)
	starts := map[string]StartEventInfo{}
	if data, err := os.ReadFile(bpmnPath); err == nil {
		if evs, err := ExtractBPMNStartEvents(string(data)); err == nil {
			for _, e := range evs {
				starts[e.ID] = e
			}
		}
	}
	triggers := make([]model.ProcessTrigger, 0, len(req.Triggers))
	for _, t := range req.Triggers {
		elementID := strings.TrimSpace(t.BPMNElementID)
		if elementID == "" {
			continue
		}
		if _, ok := starts[elementID]; !ok {
			return ProcessDTO{}, Error(CodeInvalidInput, "Trigger references unknown BPMN start event: "+elementID, http.StatusBadRequest, nil)
		}
		triggers = append(triggers, triggerDTOToModel(t, elementID))
	}
	node.Meta.Triggers = triggers
	if err := fsx.WriteYAML(node.Path, node.Meta); err != nil {
		return ProcessDTO{}, Error(CodeInternalError, "Failed to write process: "+err.Error(), http.StatusInternalServerError, err)
	}
	return GetProcess(path, id)
}

func triggerDTOToModel(t ProcessTriggerDTO, elementID string) model.ProcessTrigger {
	out := model.ProcessTrigger{
		BPMNElementID: elementID,
		Type:          normalizedTriggerType(t.Type),
		Name:          strings.TrimSpace(t.Name),
		Description:   strings.TrimSpace(t.Description),
	}
	if t.Timer != nil {
		out.Timer = &model.TimerTriggerConfig{
			Cron:        strings.TrimSpace(t.Timer.Cron),
			ISODuration: strings.TrimSpace(t.Timer.ISODuration),
			ISODate:     strings.TrimSpace(t.Timer.ISODate),
			Timezone:    strings.TrimSpace(t.Timer.Timezone),
		}
	}
	if t.Message != nil {
		out.Message = &model.MessageTriggerConfig{
			EventRef:       strings.TrimSpace(t.Message.EventRef),
			Topic:          strings.TrimSpace(t.Message.Topic),
			Filter:         strings.TrimSpace(t.Message.Filter),
			CorrelationKey: strings.TrimSpace(t.Message.CorrelationKey),
		}
	}
	if t.Signal != nil {
		out.Signal = &model.SignalTriggerConfig{SignalRef: strings.TrimSpace(t.Signal.SignalRef)}
	}
	if t.Conditional != nil {
		out.Conditional = &model.ConditionalTriggerConfig{Expression: strings.TrimSpace(t.Conditional.Expression)}
	}
	return out
}

func triggerModelToDTO(t model.ProcessTrigger) ProcessTriggerDTO {
	out := ProcessTriggerDTO{
		BPMNElementID: t.BPMNElementID,
		Name:          t.Name,
		Description:   t.Description,
		Type:          normalizedTriggerType(t.Type),
		Configured:    true,
	}
	if t.Timer != nil {
		out.Timer = &TimerTriggerConfigDTO{Cron: t.Timer.Cron, ISODuration: t.Timer.ISODuration, ISODate: t.Timer.ISODate, Timezone: t.Timer.Timezone}
	}
	if t.Message != nil {
		out.Message = &MessageTriggerConfigDTO{EventRef: t.Message.EventRef, Topic: t.Message.Topic, Filter: t.Message.Filter, CorrelationKey: t.Message.CorrelationKey}
	}
	if t.Signal != nil {
		out.Signal = &SignalTriggerConfigDTO{SignalRef: t.Signal.SignalRef}
	}
	if t.Conditional != nil {
		out.Conditional = &ConditionalTriggerConfigDTO{Expression: t.Conditional.Expression}
	}
	return out
}

// mergeTriggerView produces the effective trigger list: every start event
// from the BPMN gets one row, populated either from the configured trigger
// (matched by BPMN element ID) or as an unconfigured stub with the detected
// type carried through.
func mergeTriggerView(starts []StartEventInfo, configured []model.ProcessTrigger) []ProcessTriggerDTO {
	byID := map[string]model.ProcessTrigger{}
	for _, t := range configured {
		byID[t.BPMNElementID] = t
	}
	seen := map[string]bool{}
	out := make([]ProcessTriggerDTO, 0, len(starts)+len(configured))
	for _, s := range starts {
		seen[s.ID] = true
		if cfg, ok := byID[s.ID]; ok {
			dto := triggerModelToDTO(cfg)
			dto.DetectedType = s.Type
			if dto.Name == "" {
				dto.Name = s.Name
			}
			out = append(out, dto)
			continue
		}
		out = append(out, ProcessTriggerDTO{
			BPMNElementID: s.ID,
			Name:          s.Name,
			Type:          s.Type,
			DetectedType:  s.Type,
			Configured:    false,
		})
	}
	for _, t := range configured {
		if seen[t.BPMNElementID] {
			continue
		}
		dto := triggerModelToDTO(t)
		out = append(out, dto)
	}
	return out
}

func normalizedTriggerType(t string) string {
	switch strings.TrimSpace(strings.ToLower(t)) {
	case TriggerTypeTimer:
		return TriggerTypeTimer
	case TriggerTypeMessage:
		return TriggerTypeMessage
	case TriggerTypeSignal:
		return TriggerTypeSignal
	case TriggerTypeConditional:
		return TriggerTypeConditional
	default:
		return TriggerTypeNone
	}
}

// validateTriggers turns trigger-related problems into findings. It assumes
// the caller has already populated p.Triggers via mergeTriggerView.
func validateTriggers(p ProcessDTO, starts []StartEventInfo, add func(code, severity, msg string)) {
	startByID := map[string]StartEventInfo{}
	for _, s := range starts {
		startByID[s.ID] = s
	}
	for _, t := range p.Triggers {
		if !t.Configured {
			if t.DetectedType != TriggerTypeNone && t.DetectedType != "" {
				add("TRIGGER_TYPED_START_UNCONFIGURED", "warning", "Typed start event has no trigger configuration: "+firstNonEmpty(t.Name, t.BPMNElementID))
			}
			continue
		}
		s, known := startByID[t.BPMNElementID]
		if !known {
			add("TRIGGER_UNKNOWN_ELEMENT", "error", "Trigger references unknown BPMN start event: "+t.BPMNElementID)
			continue
		}
		if t.Type != s.Type {
			add("TRIGGER_TYPE_MISMATCH", "error", fmt.Sprintf("Trigger type %q does not match BPMN start event %q (detected %q)", t.Type, t.BPMNElementID, s.Type))
		}
		switch t.Type {
		case TriggerTypeTimer:
			validateTimerTrigger(t, add)
		case TriggerTypeMessage:
			if t.Message == nil || strings.TrimSpace(t.Message.EventRef) == "" {
				add("TRIGGER_MESSAGE_NO_REF", "warning", "Message trigger has no event_ref: "+t.BPMNElementID)
			}
		case TriggerTypeSignal:
			if t.Signal == nil || strings.TrimSpace(t.Signal.SignalRef) == "" {
				add("TRIGGER_SIGNAL_NO_REF", "warning", "Signal trigger has no signal_ref: "+t.BPMNElementID)
			}
		case TriggerTypeConditional:
			if t.Conditional == nil || strings.TrimSpace(t.Conditional.Expression) == "" {
				add("TRIGGER_CONDITIONAL_NO_EXPR", "warning", "Conditional trigger has no expression: "+t.BPMNElementID)
			}
		}
	}
}

func validateTimerTrigger(t ProcessTriggerDTO, add func(code, severity, msg string)) {
	if t.Timer == nil {
		add("TRIGGER_TIMER_NO_SCHEDULE", "warning", "Timer trigger has no schedule: "+t.BPMNElementID)
		return
	}
	cron := strings.TrimSpace(t.Timer.Cron)
	dur := strings.TrimSpace(t.Timer.ISODuration)
	date := strings.TrimSpace(t.Timer.ISODate)
	if cron == "" && dur == "" && date == "" {
		add("TRIGGER_TIMER_NO_SCHEDULE", "warning", "Timer trigger has no schedule: "+t.BPMNElementID)
		return
	}
	if cron != "" {
		if err := ValidateCronExpression(cron); err != nil {
			add("TRIGGER_TIMER_INVALID_CRON", "error", "Timer cron is invalid for "+t.BPMNElementID+": "+err.Error())
		}
	}
}

// ValidateCronExpression performs a syntactic check on a cron expression.
// It accepts:
//   - 5-field UNIX cron: minute hour day-of-month month day-of-week
//   - 6-field Quartz-style cron: second minute hour day-of-month month day-of-week
//   - macros: @yearly, @annually, @monthly, @weekly, @daily, @midnight, @hourly
//
// The check is syntactic only: it does not compute the next fire time.
func ValidateCronExpression(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return fmt.Errorf("empty cron expression")
	}
	if strings.HasPrefix(expr, "@") {
		switch strings.ToLower(expr) {
		case "@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly":
			return nil
		default:
			return fmt.Errorf("unknown cron macro %q", expr)
		}
	}
	fields := strings.Fields(expr)
	var bounds [][2]int
	switch len(fields) {
	case 5:
		bounds = [][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	case 6:
		bounds = [][2]int{{0, 59}, {0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	default:
		return fmt.Errorf("expected 5 or 6 fields, got %d", len(fields))
	}
	for i, f := range fields {
		if err := validateCronField(f, bounds[i][0], bounds[i][1]); err != nil {
			return fmt.Errorf("field %d (%q): %w", i+1, f, err)
		}
	}
	return nil
}

func validateCronField(field string, min, max int) error {
	if field == "" {
		return fmt.Errorf("empty field")
	}
	for _, part := range strings.Split(field, ",") {
		if err := validateCronToken(part, min, max); err != nil {
			return err
		}
	}
	return nil
}

func validateCronToken(token string, min, max int) error {
	step := 0
	rangePart := token
	if idx := strings.Index(token, "/"); idx >= 0 {
		var err error
		step, err = strconv.Atoi(token[idx+1:])
		if err != nil || step <= 0 {
			return fmt.Errorf("invalid step %q", token[idx+1:])
		}
		rangePart = token[:idx]
	}
	if rangePart == "*" {
		if step > 0 && step > max {
			return fmt.Errorf("step %d exceeds max %d", step, max)
		}
		return nil
	}
	if idx := strings.Index(rangePart, "-"); idx >= 0 {
		lo, err := strconv.Atoi(rangePart[:idx])
		if err != nil {
			return fmt.Errorf("invalid range start %q", rangePart[:idx])
		}
		hi, err := strconv.Atoi(rangePart[idx+1:])
		if err != nil {
			return fmt.Errorf("invalid range end %q", rangePart[idx+1:])
		}
		if lo < min || hi > max || lo > hi {
			return fmt.Errorf("range %d-%d outside [%d,%d]", lo, hi, min, max)
		}
		return nil
	}
	n, err := strconv.Atoi(rangePart)
	if err != nil {
		return fmt.Errorf("invalid number %q", rangePart)
	}
	if n < min || n > max {
		return fmt.Errorf("value %d outside [%d,%d]", n, min, max)
	}
	return nil
}
