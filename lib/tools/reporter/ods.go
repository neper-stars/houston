package reporter

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	ods "github.com/AlexJarrah/go-ods"
)

// ODSDocument wraps an ODS file for easier manipulation.
type ODSDocument struct {
	data  ods.ODS
	files *zip.ReadCloser
}

// LoadBytes loads an ODS document from bytes.
func LoadBytes(data []byte) (*ODSDocument, error) {
	odsData, files, err := ods.ReadFrom(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ODS data: %w", err)
	}
	return &ODSDocument{data: odsData, files: files}, nil
}

// LoadReader loads an ODS document from an io.Reader.
func LoadReader(r io.Reader, size int64) (*ODSDocument, error) {
	odsData, files, err := ods.ReadFrom(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ODS data: %w", err)
	}
	return &ODSDocument{data: odsData, files: files}, nil
}

// LoadFile loads an existing ODS file.
func LoadFile(filepath string) (*ODSDocument, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return LoadBytes(data)
}

// WriteBytes returns the ODS document as bytes.
func (d *ODSDocument) WriteBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := ods.WriteTo(buf, d.data, d.files); err != nil {
		return nil, fmt.Errorf("failed to write ODS: %w", err)
	}
	return buf.Bytes(), nil
}

// WriteToWriter writes the ODS document to an io.Writer.
func (d *ODSDocument) WriteToWriter(w io.Writer) error {
	return ods.WriteTo(w, d.data, d.files)
}

// Save writes the ODS document to a file.
func (d *ODSDocument) Save(filepath string) error {
	data, err := d.WriteBytes()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// Close closes the underlying zip reader.
func (d *ODSDocument) Close() error {
	if d.files != nil {
		return d.files.Close()
	}
	return nil
}

// SheetByName finds a sheet by name.
func (d *ODSDocument) SheetByName(name string) *ods.Table {
	for i := range d.data.Content.Body.Spreadsheet.Table {
		if d.data.Content.Body.Spreadsheet.Table[i].Name == name {
			return &d.data.Content.Body.Spreadsheet.Table[i]
		}
	}
	return nil
}

// SheetNames returns all sheet names.
func (d *ODSDocument) SheetNames() []string {
	var names []string
	for _, sheet := range d.data.Content.Body.Spreadsheet.Table {
		names = append(names, sheet.Name)
	}
	return names
}

// EnsureRows ensures the sheet has at least n rows.
func (d *ODSDocument) EnsureRows(sheet *ods.Table, n int) {
	for len(sheet.TableRow) < n {
		sheet.TableRow = append(sheet.TableRow, ods.TableRow{})
	}
}

// EnsureCells ensures the row has at least n cells.
func (d *ODSDocument) EnsureCells(row *ods.TableRow, n int) {
	for len(row.TableCell) < n {
		row.TableCell = append(row.TableCell, ods.TableCell{})
	}
}

// SetCellString sets a cell to a string value.
func (d *ODSDocument) SetCellString(sheet *ods.Table, row, col int, value string) {
	d.EnsureRows(sheet, row+1)
	d.EnsureCells(&sheet.TableRow[row], col+1)
	sheet.TableRow[row].TableCell[col] = ods.TableCell{
		ValueType: "string",
		P:         value,
	}
}

// SetCellInt sets a cell to an integer value.
func (d *ODSDocument) SetCellInt(sheet *ods.Table, row, col int, value int64) {
	d.EnsureRows(sheet, row+1)
	d.EnsureCells(&sheet.TableRow[row], col+1)
	sheet.TableRow[row].TableCell[col] = ods.TableCell{
		ValueType: "float",
		Value:     strconv.FormatInt(value, 10),
		P:         strconv.FormatInt(value, 10),
	}
}

// SetCellFloat sets a cell to a float value.
func (d *ODSDocument) SetCellFloat(sheet *ods.Table, row, col int, value float64) {
	d.EnsureRows(sheet, row+1)
	d.EnsureCells(&sheet.TableRow[row], col+1)
	formatted := strconv.FormatFloat(value, 'f', 2, 64)
	sheet.TableRow[row].TableCell[col] = ods.TableCell{
		ValueType: "float",
		Value:     formatted,
		P:         formatted,
	}
}

// SetCellFormula sets a cell to a formula.
// The formula should be in OpenFormula format without the "of:=" prefix.
// Example: "SUM([.B2:.B10])" for summing B2 to B10.
func (d *ODSDocument) SetCellFormula(sheet *ods.Table, row, col int, formula string) {
	d.EnsureRows(sheet, row+1)
	d.EnsureCells(&sheet.TableRow[row], col+1)
	sheet.TableRow[row].TableCell[col] = ods.TableCell{
		ValueType: "float",
		Formula:   "of:=" + formula,
		Value:     "0", // Placeholder, will be calculated when opened
		P:         "",
	}
}

// GetCellString reads a string value from a cell.
func (d *ODSDocument) GetCellString(sheet *ods.Table, row, col int) string {
	if row >= len(sheet.TableRow) {
		return ""
	}
	if col >= len(sheet.TableRow[row].TableCell) {
		return ""
	}
	return sheet.TableRow[row].TableCell[col].P
}

// GetCellInt reads an integer value from a cell.
func (d *ODSDocument) GetCellInt(sheet *ods.Table, row, col int) (int64, bool) {
	if row >= len(sheet.TableRow) {
		return 0, false
	}
	if col >= len(sheet.TableRow[row].TableCell) {
		return 0, false
	}
	cell := sheet.TableRow[row].TableCell[col]
	if cell.Value == "" {
		return 0, false
	}
	val, err := strconv.ParseInt(cell.Value, 10, 64)
	if err != nil {
		// Try parsing P field as fallback
		val, err = strconv.ParseInt(cell.P, 10, 64)
		if err != nil {
			return 0, false
		}
	}
	return val, true
}

// RowCount returns the number of rows in a sheet (including empty rows).
func (d *ODSDocument) RowCount(sheet *ods.Table) int {
	return len(sheet.TableRow)
}

// ClearSheet removes all rows from a sheet (keeping headers if specified).
func (d *ODSDocument) ClearSheet(sheet *ods.Table, keepHeaderRows int) {
	if len(sheet.TableRow) > keepHeaderRows {
		sheet.TableRow = sheet.TableRow[:keepHeaderRows]
	}
}

// AppendRow appends values as a new row to the sheet.
// Values can be strings or int64.
func (d *ODSDocument) AppendRow(sheet *ods.Table, values ...interface{}) {
	row := ods.TableRow{}
	for _, v := range values {
		var cell ods.TableCell
		switch val := v.(type) {
		case string:
			cell = ods.TableCell{
				ValueType: "string",
				P:         val,
			}
		case int:
			s := strconv.Itoa(val)
			cell = ods.TableCell{
				ValueType: "float",
				Value:     s,
				P:         s,
			}
		case int64:
			s := strconv.FormatInt(val, 10)
			cell = ods.TableCell{
				ValueType: "float",
				Value:     s,
				P:         s,
			}
		case float64:
			s := strconv.FormatFloat(val, 'f', 2, 64)
			cell = ods.TableCell{
				ValueType: "float",
				Value:     s,
				P:         s,
			}
		default:
			cell = ods.TableCell{
				ValueType: "string",
				P:         fmt.Sprintf("%v", v),
			}
		}
		row.TableCell = append(row.TableCell, cell)
	}
	sheet.TableRow = append(sheet.TableRow, row)
}

// SetHeaderRow sets the first row as headers.
func (d *ODSDocument) SetHeaderRow(sheet *ods.Table, headers ...string) {
	d.EnsureRows(sheet, 1)
	sheet.TableRow[0] = ods.TableRow{}
	for _, h := range headers {
		sheet.TableRow[0].TableCell = append(sheet.TableRow[0].TableCell, ods.TableCell{
			ValueType: "string",
			P:         h,
		})
	}
}
