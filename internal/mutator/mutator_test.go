package mutator

import (
	"bytes"
	"strings"
	"testing"
)

// --- BufferOverflowMutator ---

func TestBufferOverflowMutator_Generate_Size(t *testing.T) {
	size := 512
	m := NewBufferOverflowMutator(size)
	payload := m.Generate()
	if len(payload) != size {
		t.Errorf("expected payload length %d, got %d", size, len(payload))
	}
}

func TestBufferOverflowMutator_Generate_Content(t *testing.T) {
	m := NewBufferOverflowMutator(64)
	payload := m.Generate()
	if !bytes.Equal(payload, bytes.Repeat([]byte("A"), 64)) {
		t.Errorf("expected payload to be all 'A' bytes, got %q", payload)
	}
}

func TestBufferOverflowMutator_Generate_ZeroSize(t *testing.T) {
	m := NewBufferOverflowMutator(0)
	payload := m.Generate()
	if len(payload) != 0 {
		t.Errorf("expected empty payload, got length %d", len(payload))
	}
}

func TestBufferOverflowMutator_Name(t *testing.T) {
	m := NewBufferOverflowMutator(1)
	if m.Name() != "BufferOverflow" {
		t.Errorf("unexpected name: %q", m.Name())
	}
}

// --- FormatStringMutator ---

func TestFormatStringMutator_Generate_RepeatCount(t *testing.T) {
	specifiers := 10
	attackType := "%x."
	m := NewFormatStringMutator(specifiers, attackType)
	payload := m.Generate()
	expected := strings.Repeat(attackType, specifiers)
	if string(payload) != expected {
		t.Errorf("expected %q, got %q", expected, string(payload))
	}
}

func TestFormatStringMutator_Generate_Empty(t *testing.T) {
	m := NewFormatStringMutator(0, "%x.")
	payload := m.Generate()
	if len(payload) != 0 {
		t.Errorf("expected empty payload for 0 specifiers, got %d bytes", len(payload))
	}
}

func TestFormatStringMutator_Generate_PercentN(t *testing.T) {
	m := NewFormatStringMutator(5, "%n")
	payload := m.Generate()
	if string(payload) != "%n%n%n%n%n" {
		t.Errorf("unexpected payload: %q", string(payload))
	}
}

func TestFormatStringMutator_Name(t *testing.T) {
	m := NewFormatStringMutator(1, "%x")
	if m.Name() != "FormatString" {
		t.Errorf("unexpected name: %q", m.Name())
	}
}
