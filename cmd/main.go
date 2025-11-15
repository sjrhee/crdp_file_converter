package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"crdp-file-converter/pkg/converter"
)

// CLI flags
var (
	delimiter  string
	column     int
	encode     bool
	decode     bool
	output     string
	host       string
	port       int
	policy     string
	batchSize  int
	skipHeader bool
	timeout    int
)

// rootCmd is the entry point for the CLI application
var rootCmd = &cobra.Command{
	Use:   "crdp-file-converter <input_file>",
	Short: "CRDP File Converter - Protects/Reveals CSV/TSV columns",
	Long: `CRDP Dump File Converter

Converts CSV/TSV files by encoding/decoding specific columns using CRDP API.

Example:
  crdp-file-converter data.csv --column 1 --encode
  crdp-file-converter data.tsv --delimiter '\t' --column 2 --decode --skip-header`,
	Args: cobra.ExactArgs(1),
	Run:  runConversion,
}

// runConversion executes the file conversion process
func runConversion(cmd *cobra.Command, args []string) {
	inputFile := args[0]

	// Validate operation flags
	if err := validateOperationFlags(); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	// Determine operation
	operation := "protect"
	if decode {
		operation = "reveal"
	}

	// Auto-detect header if -s flag is not set
	if !skipHeader {
		hasHeader, err := converter.DetectHeaderLine(inputFile, delimiter, column)
		if err != nil {
			log.Fatalf("❌ Error detecting header: %v", err)
		}

		if hasHeader {
			// Ask user for confirmation
			skipHeader = promptSkipHeader()
		}
	}

	// Generate output file path if not specified
	if output == "" {
		var err error
		output, err = generateOutputPath(inputFile, encode)
		if err != nil {
			log.Fatalf("❌ Error: %v", err)
		}
	}

	log.Printf("CRDP Server: %s:%d", host, port)
	log.Printf("Policy: %s", policy)

	// Create converter and process file
	conv := converter.NewDumpConverter(host, port, policy, timeout)
	defer conv.Close()

	if err := conv.ProcessFile(
		inputFile,
		output,
		delimiter,
		column,
		operation,
		skipHeader,
		batchSize,
	); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	log.Printf("✅ Conversion completed: %s", output)
}

// promptSkipHeader asks user whether to skip the header line
// Returns true if user confirms (Y or just presses Enter)
func promptSkipHeader() bool {
	fmt.Print("Skip header line? (Y/n): ")
	
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return true // Default to skip on error
	}

	input = strings.TrimSpace(strings.ToLower(input))
	
	// Empty input (just Enter) or "y" means skip
	return input == "" || input == "y" || input == "yes"
}

// validateOperationFlags ensures exactly one of encode/decode is specified
func validateOperationFlags() error {
	if !encode && !decode {
		return fmt.Errorf("either --encode (-e) or --decode (-d) must be specified")
	}
	if encode && decode {
		return fmt.Errorf("cannot specify both --encode and --decode")
	}
	return nil
}

// generateOutputPath creates output filename with e{nn}_ or d{nn}_ prefix
// and increments counter if file already exists
func generateOutputPath(inputFile string, isEncode bool) (string, error) {
	baseName := filepath.Base(inputFile)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]

	// Determine prefix based on operation
	var prefix string
	if isEncode {
		prefix = "e01_"
	} else {
		prefix = "d01_"
	}

	// Check for duplicate names and increment if needed
	outputPath := prefix + nameWithoutExt + ext
	counter := 1

	for {
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			break
		}
		counter++
		if isEncode {
			prefix = fmt.Sprintf("e%02d_", counter)
		} else {
			prefix = fmt.Sprintf("d%02d_", counter)
		}
		outputPath = prefix + nameWithoutExt + ext
	}

	return outputPath, nil
}

func init() {
	// Disable flag sorting to maintain custom definition order
	rootCmd.Flags().SortFlags = false

	// Core operation flags
	rootCmd.Flags().BoolVarP(&encode, "encode", "e", false, "Encode (protect) data")
	rootCmd.Flags().BoolVarP(&decode, "decode", "d", false, "Decode (reveal) data")
	rootCmd.Flags().IntVarP(&column, "column", "c", -1, "Column index to convert (0-based, required)")

	// File processing flags
	rootCmd.Flags().BoolVarP(&skipHeader, "skip-header", "s", false, "Skip header line")
	rootCmd.Flags().StringVar(&delimiter, "delimiter", ",", "Column delimiter")
	rootCmd.Flags().StringVar(&output, "output", "", "Output file path (default: {e/d}{nn}_{filename}.{ext})")
	rootCmd.Flags().IntVar(&batchSize, "batch-size", 100, "Bulk API batch size")

	// CRDP server flags
	rootCmd.Flags().StringVar(&host, "host", "192.168.0.231", "CRDP host")
	rootCmd.Flags().IntVar(&port, "port", 32082, "CRDP port")
	rootCmd.Flags().StringVar(&policy, "policy", "P03", "Protection policy")
	rootCmd.Flags().IntVar(&timeout, "timeout", 5, "Request timeout in seconds")

	// Mark required flags
	rootCmd.MarkFlagRequired("column")
}

// main entry point
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
