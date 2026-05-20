package app

import (
	"strings"
	"testing"
)

func TestValidateCronExpression(t *testing.T) {
	valid := []string{
		"* * * * *",
		"0 8 * * 1",
		"*/5 * * * *",
		"0,15,30,45 * * * *",
		"0-30 9-17 * * 1-5",
		"@daily",
		"@hourly",
		"0 0 8 * * 1",
	}
	for _, e := range valid {
		if err := ValidateCronExpression(e); err != nil {
			t.Errorf("expected %q to be valid, got %v", e, err)
		}
	}
	invalid := []string{
		"",
		"* * * *",
		"60 * * * *",
		"* 24 * * *",
		"0 8 * * 1 2 3",
		"@nope",
		"*/abc * * * *",
		"5-3 * * * *",
	}
	for _, e := range invalid {
		if err := ValidateCronExpression(e); err == nil {
			t.Errorf("expected %q to be invalid", e)
		}
	}
}

func TestExtractBPMNStartEvents(t *testing.T) {
	xmlText := `<?xml version="1.0"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="P">
    <bpmn:startEvent id="StartEvent_Plain" name="Plain"/>
    <bpmn:startEvent id="StartEvent_Timer" name="Weekly">
      <bpmn:timerEventDefinition/>
    </bpmn:startEvent>
    <bpmn:startEvent id="StartEvent_Msg" name="On Event">
      <bpmn:messageEventDefinition/>
    </bpmn:startEvent>
    <bpmn:startEvent id="StartEvent_Sig"><bpmn:signalEventDefinition/></bpmn:startEvent>
    <bpmn:startEvent id="StartEvent_Cond"><bpmn:conditionalEventDefinition/></bpmn:startEvent>
  </bpmn:process>
</bpmn:definitions>`
	events, err := ExtractBPMNStartEvents(xmlText)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 5 {
		t.Fatalf("expected 5 start events, got %d: %+v", len(events), events)
	}
	want := map[string]string{
		"StartEvent_Plain": TriggerTypeNone,
		"StartEvent_Timer": TriggerTypeTimer,
		"StartEvent_Msg":   TriggerTypeMessage,
		"StartEvent_Sig":   TriggerTypeSignal,
		"StartEvent_Cond":  TriggerTypeConditional,
	}
	for _, e := range events {
		if want[e.ID] != e.Type {
			t.Errorf("event %s: want %q got %q", e.ID, want[e.ID], e.Type)
		}
	}
}

func TestProcessTriggersRoundTripAndValidation(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-TRG-001", Name: "Trigger Product"})
	if err != nil {
		t.Fatal(err)
	}
	proc, err := CreateProductProcess(p, product.ID, CreateProcessRequest{ID: "PRC-TRG-001", Name: "Trigger Process"})
	if err != nil {
		t.Fatal(err)
	}

	// Default BPMN has one plain start event "StartEvent_Begin".
	if len(proc.Triggers) != 1 {
		t.Fatalf("expected 1 default trigger row, got %d: %+v", len(proc.Triggers), proc.Triggers)
	}
	if proc.Triggers[0].Configured {
		t.Fatalf("default trigger should be unconfigured")
	}

	// Promote the start event to a timer in the BPMN XML.
	bpmnXML, err := GetProcessBPMN(p, proc.ID)
	if err != nil {
		t.Fatal(err)
	}
	timerXML := strings.Replace(
		bpmnXML,
		`<bpmn:startEvent id="StartEvent_Begin" name="Start">`,
		`<bpmn:startEvent id="StartEvent_Begin" name="Start"><bpmn:timerEventDefinition/>`,
		1,
	)
	if _, err := UpdateProcessBPMN(p, proc.ID, timerXML); err != nil {
		t.Fatal(err)
	}

	// Reject triggers pointing at unknown elements.
	if _, err := UpdateProcessTriggers(p, proc.ID, UpdateProcessTriggersRequest{
		Triggers: []ProcessTriggerDTO{{BPMNElementID: "StartEvent_Ghost", Type: TriggerTypeTimer, Timer: &TimerTriggerConfigDTO{Cron: "0 8 * * 1"}}},
	}); err == nil {
		t.Fatal("expected error for unknown BPMN element")
	}

	// Happy path: configure timer with a valid cron.
	updated, err := UpdateProcessTriggers(p, proc.ID, UpdateProcessTriggersRequest{
		Triggers: []ProcessTriggerDTO{{
			BPMNElementID: "StartEvent_Begin",
			Type:          TriggerTypeTimer,
			Timer:         &TimerTriggerConfigDTO{Cron: "0 8 * * 1"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Triggers) != 1 || !updated.Triggers[0].Configured {
		t.Fatalf("trigger not persisted: %+v", updated.Triggers)
	}
	if updated.Triggers[0].DetectedType != TriggerTypeTimer {
		t.Fatalf("detected type mismatch: %+v", updated.Triggers[0])
	}
	for _, f := range updated.Validation.Findings {
		if strings.HasPrefix(f.Code, "TRIGGER_") {
			t.Errorf("unexpected trigger finding for valid timer: %+v", f)
		}
	}

	// Invalid cron must surface as an error finding.
	bad, err := UpdateProcessTriggers(p, proc.ID, UpdateProcessTriggersRequest{
		Triggers: []ProcessTriggerDTO{{
			BPMNElementID: "StartEvent_Begin",
			Type:          TriggerTypeTimer,
			Timer:         &TimerTriggerConfigDTO{Cron: "not a cron"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(bad.Validation.Findings, "TRIGGER_TIMER_INVALID_CRON") {
		t.Fatalf("expected TRIGGER_TIMER_INVALID_CRON, got: %+v", bad.Validation.Findings)
	}

	// Type mismatch: declare message trigger for a timer start event.
	mismatch, err := UpdateProcessTriggers(p, proc.ID, UpdateProcessTriggersRequest{
		Triggers: []ProcessTriggerDTO{{
			BPMNElementID: "StartEvent_Begin",
			Type:          TriggerTypeMessage,
			Message:       &MessageTriggerConfigDTO{EventRef: "user.created"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(mismatch.Validation.Findings, "TRIGGER_TYPE_MISMATCH") {
		t.Fatalf("expected TRIGGER_TYPE_MISMATCH, got: %+v", mismatch.Validation.Findings)
	}
}

func TestTypedStartUnconfiguredFinding(t *testing.T) {
	p := createAppTestCosmos(t)
	product, err := CreateProductOffering(p, CreateProductOfferingRequest{ID: "PROD-TRG-002", Name: "Trigger Product 2"})
	if err != nil {
		t.Fatal(err)
	}
	proc, err := CreateProductProcess(p, product.ID, CreateProcessRequest{ID: "PRC-TRG-002", Name: "Trigger Process 2"})
	if err != nil {
		t.Fatal(err)
	}
	bpmnXML, err := GetProcessBPMN(p, proc.ID)
	if err != nil {
		t.Fatal(err)
	}
	timerXML := strings.Replace(
		bpmnXML,
		`<bpmn:startEvent id="StartEvent_Begin" name="Start">`,
		`<bpmn:startEvent id="StartEvent_Begin" name="Start"><bpmn:messageEventDefinition/>`,
		1,
	)
	updated, err := UpdateProcessBPMN(p, proc.ID, timerXML)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(updated.Validation.Findings, "TRIGGER_TYPED_START_UNCONFIGURED") {
		t.Fatalf("expected TRIGGER_TYPED_START_UNCONFIGURED, got: %+v", updated.Validation.Findings)
	}
}

func hasFinding(findings []FindingDTO, code string) bool {
	for _, f := range findings {
		if f.Code == code {
			return true
		}
	}
	return false
}
