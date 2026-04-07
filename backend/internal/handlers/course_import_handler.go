package handlers

import (
	"encoding/csv"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"english-learning/internal/models"
	"english-learning/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
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

func parseImportedCourseExcel(file multipart.File) ([]importedCourseRow, error) {
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer workbook.Close()

	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("spreadsheet has no sheets")
	}

	rowsData, err := workbook.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	rows := make([]importedCourseRow, 0, len(rowsData))
	seen := make(map[string]struct{})

	for index, record := range rowsData {
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

func parseImportedCourseFile(file multipart.File, filename string) ([]importedCourseRow, error) {
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension {
	case ".xlsx", ".xlsm", ".xltx", ".xltm":
		return parseImportedCourseExcel(file)
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

func buildImportedCardAudioURL(word string) *string {
	normalizedWord := strings.TrimSpace(word)
	if normalizedWord == "" {
		return nil
	}

	audioURL := fmt.Sprintf(
		"https://translate.google.com/translate_tts?ie=UTF-8&client=tw-ob&tl=en&q=%s",
		url.QueryEscape(normalizedWord),
	)

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

	rows, err := parseImportedCourseFile(file, fileHeader.Filename)
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
				buildImportedCardAudioURL(row.Word),
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
