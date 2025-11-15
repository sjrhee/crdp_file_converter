package converter

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDumpConverter(t *testing.T) {
	conv := NewDumpConverter("192.168.0.231", 32082, "P03", 10)
	if conv == nil {
		t.Fatal("NewDumpConverter returned nil")
	}
	if conv.policy != "P03" {
		t.Errorf("expected policy P03, got %s", conv.policy)
	}
}

func createTestCSV(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.WriteAll([][]string{
		{"Name", "ID", "Address"},
		{"Hong Gildong", "1234567890123", "Seoul"},
		{"Kim Chulsu", "9876543210987", "Busan"},
		{"Lee Younghee", "", "Daegu"},
	})
	w.Flush()
}

func TestProcessFile(t *testing.T) {
	t.Skip("Skipping because it requires CRDP API server")
	
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "test.csv")
	outputPath := filepath.Join(tempDir, "output.csv")

	createTestCSV(t, inputPath)

	conv := NewDumpConverter("192.168.0.231", 32082, "P03", 10)
	defer conv.Close()

	err := conv.ProcessFile(inputPath, outputPath, ",", 1, "protect", true, 100)
	if err != nil {
		t.Errorf("ProcessFile failed: %v", err)
	}

	// Check if output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}
