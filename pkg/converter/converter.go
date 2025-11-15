package converter

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"crdp-file-converter/pkg/crdp"
)

// DumpConverter handles file conversion with CRDP API
type DumpConverter struct {
	client *crdp.Client
	host   string
	port   int
	policy string
}

// NewDumpConverter creates a new DumpConverter instance with CRDP client
func NewDumpConverter(host string, port int, policy string, timeout int) *DumpConverter {
	client := crdp.NewClient(host, port, policy, timeout)
	return &DumpConverter{
		client: client,
		host:   host,
		port:   port,
		policy: policy,
	}
}

// ProcessFile orchestrates the complete file conversion workflow:
// 1. Read and parse input CSV/TSV file
// 2. Extract data to be converted
// 3. Perform bulk conversion via CRDP API
// 4. Write results to output file
func (dc *DumpConverter) ProcessFile(
	inputFile string,
	outputFile string,
	delimiter string,
	columnIndex int,
	operation string,
	skipHeader bool,
	batchSize int,
) error {
	// Validate operation
	if operation != "protect" && operation != "reveal" {
		return fmt.Errorf("operation must be 'protect' or 'reveal'")
	}

	// Validate input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", inputFile)
	}

	// Read input file and collect data to convert
	rows, dataToConvert, err := dc.readAndCollectData(inputFile, delimiter, columnIndex, skipHeader)
	if err != nil {
		return err
	}

	totalRows := len(rows)
	convertCount := len(dataToConvert)

	// Handle case where no data needs conversion
	if convertCount == 0 {
		log.Println("No data to convert.")
		return dc.writeOutput(outputFile, delimiter, rows)
	}

	log.Printf("Total %d rows, %d rows to convert (batch size: %d)", totalRows, convertCount, batchSize)

	// Perform bulk conversion
	convertedList, err := dc.performBulkConversion(operation, dataToConvert, batchSize, convertCount)
	if err != nil {
		return err
	}

	if len(convertedList) != len(dataToConvert) {
		return fmt.Errorf("result count mismatch: requested %d, got %d", len(dataToConvert), len(convertedList))
	}

	// Write converted output
	err = dc.writeConvertedOutput(outputFile, delimiter, rows, columnIndex, dataToConvert, convertedList)
	if err != nil {
		return err
	}
	
	// Report summary
	log.Printf("✅ Conversion completed: %s (%d/%d rows processed)", outputFile, totalRows, totalRows)
	
	return nil
}

// readAndCollectData reads CSV file and extracts data to be converted
// It returns:
// - rows: all rows with metadata about how they should be processed
// - dataToConvert: extracted data values for CRDP API
// - error: any read error
func (dc *DumpConverter) readAndCollectData(inputFile string, delimiter string, columnIndex int, skipHeader bool) (
	[]map[string]interface{},
	[]string,
	error,
) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.Comma = rune(delimiter[0])

	var rows []map[string]interface{}
	var dataToConvert []string
	lineNum := 0

	for {
		record, err := r.Read()
		if err != nil {
			break
		}

		lineNum++

		// Skip first line only if skipHeader flag is set
		if skipHeader && lineNum == 1 {
			rows = append(rows, map[string]interface{}{
				"type": "header",
				"row":  record,
				"data": nil,
			})
			continue
		}

		// Validate column index exists
		if columnIndex >= len(record) {
			log.Printf("Warning: line %d does not have column %d (total columns: %d)", lineNum, columnIndex, len(record))
			rows = append(rows, map[string]interface{}{
				"type": "skip",
				"row":  record,
				"data": nil,
			})
			continue
		}

		originalData := record[columnIndex]

		// Skip empty values
		if strings.TrimSpace(originalData) == "" {
			rows = append(rows, map[string]interface{}{
				"type": "empty",
				"row":  record,
				"data": nil,
			})
			continue
		}

		// Mark for conversion
		rows = append(rows, map[string]interface{}{
			"type": "convert",
			"row":  record,
			"data": originalData,
		})
		dataToConvert = append(dataToConvert, originalData)
	}

	return rows, dataToConvert, nil
}

// isNumericRow checks if the value at columnIndex looks like a number
func isNumericRow(record []string, columnIndex int) bool {
	if columnIndex >= len(record) {
		return false
	}
	val := strings.TrimSpace(record[columnIndex])
	if val == "" {
		return false
	}
	// Check if value starts with digit (likely numeric data, not a header)
	return val[0] >= '0' && val[0] <= '9'
}

// DetectHeaderLine checks if the input file appears to have a header line
// Returns true if first line looks like a header (non-numeric text)
func DetectHeaderLine(inputFile string, delimiter string, columnIndex int) (bool, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return false, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.Comma = rune(delimiter[0])

	record, err := r.Read()
	if err != nil {
		return false, err
	}

	// If first row value is not numeric, it's likely a header
	return !isNumericRow(record, columnIndex), nil
}

// performBulkConversion calls CRDP API in batches and collects converted data
func (dc *DumpConverter) performBulkConversion(operation string, dataToConvert []string, batchSize int, totalCount int) ([]string, error) {
	// Split data into batches
	batches := make([][]string, 0)
	for i := 0; i < len(dataToConvert); i += batchSize {
		end := i + batchSize
		if end > len(dataToConvert) {
			end = len(dataToConvert)
		}
		batches = append(batches, dataToConvert[i:end])
	}

	var convertedList []string
	processedCount := 0
	barWidth := 50

	for batchIdx, batchData := range batches {
		// Call appropriate CRDP API
		var resp *crdp.APIResponse
		if operation == "protect" {
			resp = dc.client.ProtectBulk(batchData)
		} else {
			resp = dc.client.RevealBulk(batchData)
		}

		// Check for errors
		if !resp.IsSuccess() {
			return nil, fmt.Errorf("batch %d API call failed: %d - %v", batchIdx+1, resp.StatusCode, resp.Body)
		}

		// Extract converted data from response
		var batchConverted []string
		if operation == "protect" {
			batchConverted = dc.client.ExtractProtectedListFromProtectResponse(resp)
		} else {
			batchConverted = dc.client.ExtractRestoredListFromRevealResponse(resp)
		}

		// Handle partial results - pad with empty strings if needed
		if len(batchConverted) != len(batchData) {
			for i := len(batchConverted); i < len(batchData); i++ {
				batchConverted = append(batchConverted, "")
			}
		}

		convertedList = append(convertedList, batchConverted...)
		processedCount += len(batchConverted)

		// Calculate progress percentage
		percent := (processedCount * 100) / totalCount
		filledChars := (processedCount * barWidth) / totalCount
		
		// Build progress bar
		progressBar := ""
		for i := 0; i < barWidth; i++ {
			if i < filledChars {
				progressBar += "█"
			} else {
				progressBar += "░"
			}
		}
		
		// Show progress
		fmt.Printf("\rProcessing: [%s] %d%% (%d/%d)", progressBar, percent, processedCount, totalCount)
	}

	// Final progress bar at 100%
	progressBar := ""
	for i := 0; i < barWidth; i++ {
		progressBar += "█"
	}
	fmt.Printf("\rProcessing: [%s] 100%% (%d/%d)\n", progressBar, totalCount, totalCount)
	return convertedList, nil
}

// writeConvertedOutput writes the conversion results to output file
func (dc *DumpConverter) writeConvertedOutput(
	outputFile string,
	delimiter string,
	rows []map[string]interface{},
	columnIndex int,
	originalData []string,
	convertedData []string,
) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.Comma = rune(delimiter[0])

	processedCount := 0
	errorCount := 0
	convertIndex := 0

	for _, rowMap := range rows {
		rowType := rowMap["type"].(string)
		row := rowMap["row"].([]string)

		switch rowType {
		case "header", "skip", "empty":
			// Write as-is
			w.Write(row)
		case "convert":
			// Replace column value with converted data
			if convertIndex >= len(convertedData) {
				errorCount++
				w.Write(row)
				convertIndex++
				continue
			}

			convertedValue := convertedData[convertIndex]
			if convertedValue == "" {
				log.Printf("Warning: converted value is empty (index: %d)", convertIndex)
				errorCount++
			} else {
				processedCount++
			}

			newRow := make([]string, len(row))
			copy(newRow, row)
			newRow[columnIndex] = convertedValue
			w.Write(newRow)
			convertIndex++
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return nil
}

// writeOutput writes rows to output file without conversion
func (dc *DumpConverter) writeOutput(outputFile string, delimiter string, rows []map[string]interface{}) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	w.Comma = rune(delimiter[0])

	for _, rowMap := range rows {
		row := rowMap["row"].([]string)
		w.Write(row)
	}

	w.Flush()
	return w.Error()
}

// Close closes the converter and underlying resources
func (dc *DumpConverter) Close() error {
	if dc.client != nil {
		return dc.client.Close()
	}
	return nil
}
