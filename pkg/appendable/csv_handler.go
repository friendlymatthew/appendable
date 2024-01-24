package appendable

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/kevmo314/appendable/pkg/protocol"
)

type CSVHandler struct {
	io.ReadSeeker
}

func (c CSVHandler) Synchronize(f *IndexFile) error {
	f.Logger.Debug("Starting CSV synchronization")

	var headers []string
	var err error

	fromNewIndexFile := false

	isHeader := false

	if len(f.Indexes) == 0 {
		isHeader = true
		fromNewIndexFile = true
	} else {
		for _, index := range f.Indexes {
			headers = append(headers, index.FieldName)
		}
	}

	scanner := bufio.NewScanner(f.data)

	for i := 0; scanner.Scan(); i++ {
		line := scanner.Bytes()

		existingCount := len(f.EndByteOffsets)

		// append a data range
		var start uint64
		if len(f.EndByteOffsets) > 0 {
			start = f.EndByteOffsets[existingCount-1]
		}

		f.EndByteOffsets = append(f.EndByteOffsets, start+uint64(len(line))+1)

		f.Checksums = append(f.Checksums, xxhash.Sum64(line))

		if isHeader {
			f.Logger.Info("Parsing CSV headers")
			dec := csv.NewReader(bytes.NewReader(line))
			headers, err = dec.Read()
			if err != nil {
				f.Logger.Error("failed to parse CSV header", "error", err)
				return fmt.Errorf("failed to parse CSV header: %w", err)
			}
			isHeader = false
			continue
		}

		dec := csv.NewReader(bytes.NewReader(line))
		f.Logger.Debug("Handling csv", "line", i)
		f.handleCSVLine(dec, headers, []string{}, uint64(existingCount)-1, start)
		f.Logger.Info("Succesfully processed", "line", i)
	}

	if fromNewIndexFile && len(f.EndByteOffsets) > 0 {
		f.EndByteOffsets = f.EndByteOffsets[1:]
		f.Checksums = f.Checksums[1:]

		f.Logger.Debug("Trimming endbyte offsets and checksums", "endByteOffsets", slog.Any("endByteOffsets", f.EndByteOffsets), "checksums", slog.Any("checksums", f.Checksums))
	}

	f.Logger.Debug("Ending CSV synchronization")
	return nil
}

func fieldRankCsvField(fieldValue any) int {
	switch fieldValue.(type) {
	case nil:
		return 1
	case bool:
		return 2
	case int, int8, int16, int32, int64, float32, float64:
		return 3
	case string:
		return 4
	default:
		panic("unknown type")
	}
}

func inferCSVField(fieldValue string) (interface{}, protocol.FieldType) {
	if fieldValue == "" {
		return nil, protocol.FieldTypeNull
	}

	if i, err := strconv.Atoi(fieldValue); err == nil {
		return i, protocol.FieldTypeNumber
	}

	if f, err := strconv.ParseFloat(fieldValue, 64); err == nil {
		return f, protocol.FieldTypeNumber
	}

	if b, err := strconv.ParseBool(fieldValue); err == nil {
		return b, protocol.FieldTypeBoolean
	}

	return fieldValue, protocol.FieldTypeString
}

func (i *IndexFile) handleCSVLine(dec *csv.Reader, headers []string, path []string, dataIndex, dataOffset uint64) error {
	i.Logger.Debug("Processing CSV line", slog.Int("dataIndex", int(dataIndex)), slog.Int("dataOffset", int(dataOffset)))

	record, err := dec.Read()

	if err != nil {
		i.Logger.Error("Failed to read CSV record at index", "dataIndex", dataIndex, "error", err)
		return fmt.Errorf("failed to read CSV record at index %d: %w", dataIndex, err)
	}

	i.Logger.Debug("CSV line read successfully", "record", record)

	cumulativeLength := uint64(0)

	for fieldIndex, fieldValue := range record {
		if fieldIndex >= len(headers) {
			i.Logger.Error("Field index is out of bounds with headers", "fieldIndex", fieldIndex, "headers", slog.Any("headers", headers))
			return fmt.Errorf("field index %d is out of bounds with header", fieldIndex)
		}

		fieldName := headers[fieldIndex]
		name := strings.Join(append(path, fieldName), ".")

		fieldOffset := dataOffset + cumulativeLength
		fieldLength := uint64(len(fieldValue))

		value, fieldType := inferCSVField(fieldValue)

		switch fieldType {
		case protocol.FieldTypeBoolean, protocol.FieldTypeString, protocol.FieldTypeNumber:
			tree := i.Indexes[i.findIndex(name, value)].IndexRecords

			tree[value] = append(tree[value], protocol.IndexRecord{
				DataNumber:           dataIndex,
				FieldStartByteOffset: uint64(fieldOffset),
				FieldLength:          int(fieldLength),
			})

			i.Logger.Debug("Appended index record",
				slog.String("field", name),
				slog.Any("value", value),
				slog.Int("start", int(fieldOffset)))

		case protocol.FieldTypeNull:
			for j := range i.Indexes {
				if i.Indexes[j].FieldName == name {
					i.Indexes[j].FieldType |= protocol.FieldTypeNull
				}
			}
			i.Logger.Debug("Marked field", "name", name)

		default:
			i.Logger.Error("Encountered unexpected type '%T' for field '%s'", value, name)
			return fmt.Errorf("unexpected type '%T'", value)
		}

		cumulativeLength += fieldLength
	}

	return nil
}
