package handlers

import (
	"archive/zip"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"english-learning/internal/models"
	"english-learning/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type importedCourseRow struct {
	Lesson      int
	Position    int
	Word        string
	Translation string
}

type importCourseResponse struct {
	Course       *models.Course `json:"course"`
	LessonsCount int            `json:"lessons_count"`
	CardsCount   int            `json:"cards_count"`
}

type xlsxWorkbook struct {
	Sheets []struct {
		Name string `xml:"name,attr"`
		ID   string `xml:"id,attr"`
		RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	} `xml:"sheets>sheet"`
}

type xlsxRelationships struct {
	Items []struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

type xlsxSharedStrings struct {
	Items []struct {
		Text string `xml:"t"`
		Runs []struct {
			Text string `xml:"t"`
		} `xml:"r"`
	} `xml:"si"`
}

type xlsxWorksheet struct {
	Rows []struct {
		Cells []struct {
			Reference string `xml:"r,attr"`
			Type      string `xml:"t,attr"`
			Value     string `xml:"v"`
			Inline    struct {
				Text string `xml:"t"`
				Runs []struct {
					Text string `xml:"t"`
				} `xml:"r"`
			} `xml:"is"`
		} `xml:"c"`
	} `xml:"sheetData>row"`
}

func normalizeImportHeaderValue(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
}

func isCourseImportHeader(columns []string) bool {
	if len(columns) < 4 {
		return false
	}

	lessonValue := normalizeImportHeaderValue(columns[0])
	positionValue := normalizeImportHeaderValue(columns[1])
	englishValue := normalizeImportHeaderValue(columns[2])
	russianValue := normalizeImportHeaderValue(columns[3])

	return (lessonValue == "урок" || lessonValue == "lesson") &&
		(positionValue == "№" || positionValue == "no" || positionValue == "n" || positionValue == "номер") &&
		(englishValue == "en" || englishValue == "english") &&
		(russianValue == "ru" || russianValue == "russian" || russianValue == "translation")
}

func spreadsheetColumnValue(columns []string, index int) string {
	if index >= len(columns) {
		return ""
	}

	return strings.TrimPrefix(strings.TrimSpace(columns[index]), "\ufeff")
}

func parseImportedCourseColumns(columns []string, rowNumber int, seen map[string]struct{}) (*importedCourseRow, bool, error) {
	lessonValue := spreadsheetColumnValue(columns, 0)
	positionValue := spreadsheetColumnValue(columns, 1)
	word := spreadsheetColumnValue(columns, 2)
	translation := spreadsheetColumnValue(columns, 3)

	if lessonValue == "" && positionValue == "" && word == "" && translation == "" {
		return nil, true, nil
	}

	if rowNumber == 1 && isCourseImportHeader([]string{lessonValue, positionValue, word, translation}) {
		return nil, true, nil
	}

	lesson, err := strconv.Atoi(lessonValue)
	if err != nil || lesson <= 0 {
		return nil, false, fmt.Errorf("row %d: lesson must be a positive integer", rowNumber)
	}

	position, err := strconv.Atoi(positionValue)
	if err != nil || position <= 0 {
		return nil, false, fmt.Errorf("row %d: number must be a positive integer", rowNumber)
	}

	if word == "" || translation == "" {
		return nil, false, fmt.Errorf("row %d: word and translation are required", rowNumber)
	}

	uniqueKey := fmt.Sprintf("%d:%d", lesson, position)
	if _, exists := seen[uniqueKey]; exists {
		return nil, false, fmt.Errorf("lesson %d has duplicate number %d", lesson, position)
	}
	seen[uniqueKey] = struct{}{}

	return &importedCourseRow{
		Lesson:      lesson,
		Position:    position,
		Word:        word,
		Translation: translation,
	}, false, nil
}

func parseImportedCourseCSV(file multipart.File, delimiter rune) ([]importedCourseRow, error) {
	reader := csv.NewReader(file)
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	rows := make([]importedCourseRow, 0, len(records))
	seen := make(map[string]struct{})

	for index, record := range records {
		row, skip, err := parseImportedCourseColumns(record, index+1, seen)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}
		rows = append(rows, *row)
	}

	return rows, nil
}

func readZipFile(reader *zip.Reader, name string) ([]byte, error) {
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		return io.ReadAll(rc)
	}

	return nil, fmt.Errorf("xlsx entry not found: %s", name)
}

func cellReferenceToIndex(reference string) int {
	if reference == "" {
		return -1
	}

	letters := make([]rune, 0, len(reference))
	for _, char := range reference {
		if char >= 'A' && char <= 'Z' {
			letters = append(letters, char)
			continue
		}
		if char >= 'a' && char <= 'z' {
			letters = append(letters, char-32)
			continue
		}
		break
	}

	if len(letters) == 0 {
		return -1
	}

	index := 0
	for _, char := range letters {
		index = index*26 + int(char-'A'+1)
	}

	return index - 1
}

func sharedStringValue(sharedStrings []string, rawValue string) string {
	index, err := strconv.Atoi(strings.TrimSpace(rawValue))
	if err != nil || index < 0 || index >= len(sharedStrings) {
		return ""
	}

	return sharedStrings[index]
}

func parseXLSXSharedStrings(reader *zip.Reader) ([]string, error) {
	data, err := readZipFile(reader, "xl/sharedStrings.xml")
	if err != nil {
		return []string{}, nil
	}

	var shared xlsxSharedStrings
	if err := xml.Unmarshal(data, &shared); err != nil {
		return nil, err
	}

	values := make([]string, 0, len(shared.Items))
	for _, item := range shared.Items {
		if item.Text != "" {
			values = append(values, item.Text)
			continue
		}

		var builder strings.Builder
		for _, run := range item.Runs {
			builder.WriteString(run.Text)
		}
		values = append(values, builder.String())
	}

	return values, nil
}

func resolveXLSXFirstSheetPath(reader *zip.Reader) (string, error) {
	workbookData, err := readZipFile(reader, "xl/workbook.xml")
	if err != nil {
		return "", err
	}

	var workbook xlsxWorkbook
	if err := xml.Unmarshal(workbookData, &workbook); err != nil {
		return "", err
	}

	if len(workbook.Sheets) == 0 {
		return "", fmt.Errorf("spreadsheet has no sheets")
	}

	relationshipsData, err := readZipFile(reader, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return "", err
	}

	var relationships xlsxRelationships
	if err := xml.Unmarshal(relationshipsData, &relationships); err != nil {
		return "", err
	}

	firstSheetRelationshipID := workbook.Sheets[0].RID
	for _, relationship := range relationships.Items {
		if relationship.ID != firstSheetRelationshipID {
			continue
		}

		target := relationship.Target
		if strings.HasPrefix(target, "/") {
			return strings.TrimPrefix(target, "/"), nil
		}

		return path.Clean(path.Join("xl", target)), nil
	}

	return "", fmt.Errorf("first sheet relationship not found")
}

func parseImportedCourseExcel(file multipart.File, size int64) ([]importedCourseRow, error) {
	reader, err := zip.NewReader(file, size)
	if err != nil {
		return nil, err
	}

	sharedStrings, err := parseXLSXSharedStrings(reader)
	if err != nil {
		return nil, err
	}

	sheetPath, err := resolveXLSXFirstSheetPath(reader)
	if err != nil {
		return nil, err
	}

	sheetData, err := readZipFile(reader, sheetPath)
	if err != nil {
		return nil, err
	}

	var worksheet xlsxWorksheet
	if err := xml.Unmarshal(sheetData, &worksheet); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	rows := make([]importedCourseRow, 0, len(worksheet.Rows))

	for rowIndex, row := range worksheet.Rows {
		columns := make([]string, 4)
		for _, cell := range row.Cells {
			index := cellReferenceToIndex(cell.Reference)
			if index < 0 || index > 3 {
				continue
			}

			value := cell.Value
			switch cell.Type {
			case "s":
				value = sharedStringValue(sharedStrings, cell.Value)
			case "inlineStr":
				if cell.Inline.Text != "" {
					value = cell.Inline.Text
				} else {
					var builder strings.Builder
					for _, run := range cell.Inline.Runs {
						builder.WriteString(run.Text)
					}
					value = builder.String()
				}
			}

			columns[index] = strings.TrimSpace(value)
		}

		parsedRow, skip, err := parseImportedCourseColumns(columns, rowIndex+1, seen)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}

		rows = append(rows, *parsedRow)
	}

	return rows, nil
}

func parseImportedCourseFile(file multipart.File, size int64, filename string) ([]importedCourseRow, error) {
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension {
	case ".xlsx":
		return parseImportedCourseExcel(file, size)
	case ".csv":
		return parseImportedCourseCSV(file, ',')
	case ".tsv", ".txt":
		return parseImportedCourseCSV(file, '\t')
	default:
		return nil, fmt.Errorf("unsupported file format: use .xlsx, .csv or .tsv")
	}
}

func sortImportedCourseRows(rows []importedCourseRow) {
	sort.Slice(rows, func(leftIndex, rightIndex int) bool {
		left := rows[leftIndex]
		right := rows[rightIndex]

		if left.Lesson != right.Lesson {
			return left.Lesson < right.Lesson
		}

		return left.Position < right.Position
	})
}

func (h *CourseHandler) buildImportedCardAudioURL(word string) *string {
	normalizedWord := strings.TrimSpace(strings.ToLower(word))
	if normalizedWord == "" || h.dictionaryService == nil {
		return nil
	}

	audioURL, err := h.dictionaryService.GetWordAudioURL(normalizedWord)
	if err != nil || strings.TrimSpace(audioURL) == "" {
		return nil
	}

	return &audioURL
}

func parseOptionalBool(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true") || strings.TrimSpace(value) == "1"
}

func getAuthenticatedUserID(c *gin.Context) *uuid.UUID {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	userIDString, ok := userIDValue.(string)
	if !ok || strings.TrimSpace(userIDString) == "" {
		return nil
	}

	parsedUserID, err := uuid.Parse(userIDString)
	if err != nil {
		return nil
	}

	return &parsedUserID
}

func (h *CourseHandler) CreateCourseFromImport(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	slug := strings.TrimSpace(c.PostForm("slug"))

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	fileHeader, err := c.FormFile("import_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "import_file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open import file"})
		return
	}
	defer file.Close()

	rows, err := parseImportedCourseFile(file, fileHeader.Size, fileHeader.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no importable rows found in file"})
		return
	}

	sortImportedCourseRows(rows)

	var description *string
	if value := strings.TrimSpace(c.PostForm("description")); value != "" {
		description = &value
	}

	var imageURL *string
	if value := strings.TrimSpace(c.PostForm("image_url")); value != "" {
		imageURL = &value
	}

	createdBy := getAuthenticatedUserID(c)

	course, err := h.courseService.CreateCourse(
		title,
		slug,
		description,
		imageURL,
		parseOptionalBool(c.PostForm("is_published")),
		createdBy,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	failImport := func(importErr error) {
		if deleteErr := h.courseService.DeleteCourse(course.ID); deleteErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("%s (cleanup failed: %s)", importErr.Error(), deleteErr.Error()),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": importErr.Error()})
	}

	lessonRows := make(map[int][]importedCourseRow)
	for _, row := range rows {
		lessonRows[row.Lesson] = append(lessonRows[row.Lesson], row)
	}

	lessonNumbers := make([]int, 0, len(lessonRows))
	for lessonNumber := range lessonRows {
		lessonNumbers = append(lessonNumbers, lessonNumber)
	}
	sort.Ints(lessonNumbers)

	for _, lessonNumber := range lessonNumbers {
		deck, err := h.deckService.CreateDeck(
			course.ID,
			fmt.Sprintf("Урок %d", lessonNumber),
			utils.Slugify(fmt.Sprintf("lesson-%d", lessonNumber), "lesson"),
			nil,
			lessonNumber,
		)
		if err != nil {
			failImport(err)
			return
		}

		lessonCards := lessonRows[lessonNumber]
		sort.Slice(lessonCards, func(leftIndex, rightIndex int) bool {
			return lessonCards[leftIndex].Position < lessonCards[rightIndex].Position
		})

		for _, row := range lessonCards {
			position := row.Position
			if _, err := h.cardService.CreateCard(
				deck.ID,
				&position,
				row.Word,
				row.Translation,
				nil,
				h.buildImportedCardAudioURL(row.Word),
				nil,
				nil,
				createdBy,
				false,
			); err != nil {
				failImport(err)
				return
			}
		}
	}

	c.JSON(http.StatusCreated, importCourseResponse{
		Course:       course,
		LessonsCount: len(lessonNumbers),
		CardsCount:   len(rows),
	})
}
