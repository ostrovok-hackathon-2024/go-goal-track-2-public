package coolcsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// Reader represents a CSV reader
type Reader struct {
	reader *csv.Reader
	file   *os.File
}

// Writer represents a CSV writer
type Writer struct {
	writer *csv.Writer
	file   *os.File
}

// NewReader creates a new CSV reader
func NewReader(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	return &Reader{
		reader: csv.NewReader(file),
		file:   file,
	}, nil
}

// NewWriter creates a new CSV writer
func NewWriter(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	return &Writer{
		writer: csv.NewWriter(file),
		file:   file,
	}, nil
}

// ReadAll reads all records from the CSV file
func (r *Reader) ReadAll() ([][]string, error) {
	records, err := r.reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}
	return records, nil
}

// Read reads one record from the CSV file
func (r *Reader) Read() ([]string, error) {
	record, err := r.reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("error reading CSV record: %w", err)
	}
	return record, nil
}

// Close closes the CSV reader
func (r *Reader) Close() error {
	return r.file.Close()
}

// WriteAll writes all records to the CSV file
func (w *Writer) WriteAll(records [][]string) error {
	err := w.writer.WriteAll(records)
	if err != nil {
		return fmt.Errorf("error writing CSV: %w", err)
	}
	return nil
}

// Write writes one record to the CSV file
func (w *Writer) Write(record []string) error {
	err := w.writer.Write(record)
	if err != nil {
		return fmt.Errorf("error writing CSV record: %w", err)
	}
	return nil
}

// Flush writes any buffered data to the underlying io.Writer
func (w *Writer) Flush() {
	w.writer.Flush()
}

// Close flushes the writer and closes the CSV writer
func (w *Writer) Close() error {
	w.Flush()
	return w.file.Close()
}
